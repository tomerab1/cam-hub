import { createContext, useContext, useEffect, useMemo, useState } from "react";
import type { CameraDto } from "../contracts/CameraDto";

type Ctx = {
	cameras: CameraDto[];
	loading: boolean;
	error: string | null;
	refresh: () => void;
	upsert: (cam: Partial<CameraDto> & { uuid: string }) => void;
};

const CameraDtosCtx = createContext<Ctx | null>(null);

export function CameraProvider({ children }: { children: React.ReactNode }) {
	const [cameras, setCameraDtos] = useState<CameraDto[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [tick, setTick] = useState(0);

	useEffect(() => {
		(async () => {
			try {
				setLoading(true);
				setError(null);
				const res = await fetch(
					"http://localhost:5555/api/v1/cameras?offset=0&limit=100"
				);
				if (!res.ok) throw new Error(`HTTP ${res.status}`);
				const data: CameraDto[] = await res.json();
				setCameraDtos(data);
			} catch (e: unknown) {
				console.error(e);
				if (e instanceof Error) setError(e.message ?? "fetch failed");
				else setError("Unknown error");
			} finally {
				setLoading(false);
			}
		})();
	}, [tick]);

	const upsert = (cam: Partial<CameraDto> & { uuid: string }) => {
		setCameraDtos((prev) => {
			const i = prev.findIndex((c) => c.uuid === cam.uuid);
			if (i === -1)
				return [
					{
						uuid: cam.uuid,
						camera_name: cam.camera_name ?? "(unnamed)",
						manufacturer: cam.manufacturer ?? "",
						model: cam.model ?? "",
						firmware_version: cam.firmware_version ?? "",
						serial_number: cam.serial_number ?? "",
						hardware_id: cam.hardware_id ?? "",
						addr: cam.addr ?? "",
						is_paired: cam.is_paired ?? false,
					},
					...prev,
				];

			const copy = prev.slice();
			copy[i] = { ...copy[i], ...cam };
			return copy;
		});
	};

	useEffect(() => {
		const es = new EventSource("http://localhost:5555/api/v1/events/discovery");
		es.onmessage = (e) => {
			try {
				const evt = JSON.parse(e.data);
				if (evt.type === "device_new") {
					upsert({ uuid: evt.uuid, addr: evt.addr });
				} else if (evt.type === "device_ip_changed") {
					upsert({ uuid: evt.uuid, addr: evt.addr });
				}
			} catch (err) {
				console.error(err);
			}
		};
		return () => es.close();
	}, []);

	const value = useMemo<Ctx>(
		() => ({
			cameras,
			loading,
			error,
			refresh: () => setTick((t) => t + 1),
			upsert,
		}),
		[cameras, loading, error]
	);

	return (
		<CameraDtosCtx.Provider value={value}>{children}</CameraDtosCtx.Provider>
	);
}

export function useCameras() {
	const ctx = useContext(CameraDtosCtx);
	if (!ctx)
		throw new Error("useCameraDtos must be used within CameraDtosProvider");
	return ctx;
}

import {
	createContext,
	useCallback,
	useContext,
	useEffect,
	useMemo,
	useState,
} from "react";
import type { CameraDto } from "../contracts/CameraDto";

type Ctx = {
	cameras: CameraDto[];
	loading: boolean;
	error: string | null;
	refresh: () => void;
	upsert: (
		cam: Partial<CameraDto> & { uuid: string; whepUrl: string | null }
	) => void;
	getStreamUrl: (uuid: string) => Promise<string | null>;
	invalidateStreamUrl: (uuid: string) => void;
	deleteStream: (uuid: string) => void;
};

const CameraDtosCtx = createContext<Ctx | null>(null);

export function CameraProvider({ children }: { children: React.ReactNode }) {
	const [cameras, setCameraDtos] = useState<CameraDto[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [tick, setTick] = useState(0);

	const streamCache = useMemo(
		() => new Map<string, { url: string | null; ts: number }>(),
		[]
	);
	const inflight = useMemo(() => new Map<string, Promise<string | null>>(), []);
	const STREAM_TTL_MS = 60_000;

	useEffect(() => {
		(async () => {
			try {
				setLoading(true);
				setError(null);
				const res = await fetch(
					"http://localhost:5555/api/v1/cameras?offset=0&limit=100"
				);
				if (!res.ok) {
					throw new Error(`HTTP ${res.status}`);
				}
				const data: CameraDto[] = await res.json();
				setCameraDtos(data);
			} catch (e: unknown) {
				console.error(e);
				if (e instanceof Error) {
					setError(e.message ?? "fetch failed");
				} else {
					setError("Unknown error");
				}
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

	const deleteStream = useCallback(async (uuid: string) => {
		const ctrl = new AbortController();
		try {
			const resp = await fetch(
				`http://localhost:5555/api/v1/cameras/${uuid}/stream`,
				{
					method: "DELETE",
					signal: ctrl.signal,
					keepalive: true,
				}
			);
			if (!resp.ok || resp.status !== 404) {
				throw new Error(`HTTP ${resp.status}`);
			}
		} catch (err) {
			console.error(err);
		}
	}, []);

	const getStreamUrl = useCallback(
		async (uuid: string): Promise<string | null> => {
			const now = Date.now();
			const cached = streamCache.get(uuid);
			if (cached && now - cached.ts < STREAM_TTL_MS) return cached.url;

			const existing = inflight.get(uuid);
			if (existing) return existing;

			const p = (async () => {
				try {
					const resp = await fetch(
						`http://localhost:5555/api/v1/cameras/${uuid}/stream`
					);
					if (!resp.ok) {
						throw new Error(`HTTP ${resp.status}`);
					}

					const data: { url: string | null } = await resp.json();
					streamCache.set(uuid, { url: data.url, ts: Date.now() });

					return data.url;
				} catch (err) {
					console.error(err);
					streamCache.set(uuid, { url: null, ts: Date.now() });

					return null;
				} finally {
					inflight.delete(uuid);
				}
			})();

			inflight.set(uuid, p);
			return p;
		},
		[inflight, streamCache]
	);

	const invalidateStreamUrl = useCallback(
		(uuid: string) => {
			streamCache.delete(uuid);
			inflight.delete(uuid);
		},
		[inflight, streamCache]
	);

	useEffect(() => {
		const es = new EventSource("http://localhost:5555/api/v1/events/discovery");
		es.onmessage = (e) => {
			try {
				const evt = JSON.parse(e.data);
				if (evt.type === "device_new") {
					upsert({ uuid: evt.uuid, addr: evt.addr });
					invalidateStreamUrl(evt.uuid);
				} else if (evt.type === "device_ip_changed") {
					upsert({ uuid: evt.uuid, addr: evt.addr });
					deleteStream(evt.uuid);
					invalidateStreamUrl(evt.uuid);
				}
			} catch (err) {
				console.error(err);
			}
		};
		return () => es.close();
	}, [invalidateStreamUrl, deleteStream]);

	const value = useMemo<Ctx>(
		() => ({
			cameras,
			loading,
			error,
			refresh: () => setTick((t) => t + 1),
			upsert,
			getStreamUrl,
			invalidateStreamUrl,
			deleteStream,
		}),
		[cameras, loading, error, getStreamUrl, invalidateStreamUrl, deleteStream]
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

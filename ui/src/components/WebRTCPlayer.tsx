import { useEffect, useRef, useState } from "react";

interface WebRTCProps {
	whepUrl: string;
	onPlayError?: (err: Error) => void;
	stunServers?: RTCIceServer[];
	autoRetry?: boolean;
	retryDelayMs?: number;
	onVideoDims?: (w: number, h: number) => void;
}

async function waitForIceComplete(pc: RTCPeerConnection, timeoutMs = 4000) {
	if (pc.iceGatheringState === "complete") return;
	await Promise.race([
		new Promise<void>((resolve) => {
			const h = () => {
				if (pc.iceGatheringState === "complete") {
					pc.removeEventListener("icegatheringstatechange", h);
					resolve();
				}
			};
			pc.addEventListener("icegatheringstatechange", h);
		}),
		new Promise<void>((_, reject) =>
			setTimeout(() => reject(new Error("ICE gathering timeout")), timeoutMs)
		),
	]);
}

export default function WebRTCPlayer({
	whepUrl,
	onPlayError,
	stunServers,
	autoRetry = true,
	retryDelayMs = 1000,
	onVideoDims,
}: WebRTCProps) {
	const videoRef = useRef<HTMLVideoElement>(null);
	const pcRef = useRef<RTCPeerConnection | null>(null);
	const resourceUrlRef = useRef<string | null>(null);
	const fetchAbortRef = useRef<AbortController | null>(null);
	const [status, setStatus] = useState<
		"idle" | "connecting" | "playing" | "error"
	>("idle");

	useEffect(() => {
		let cancelled = false;

		const setup = async () => {
			cleanup();
			setStatus("connecting");

			const pc = new RTCPeerConnection({
				iceServers: stunServers ?? [{ urls: "stun:stun.l.google.com:19302" }],
			});
			pcRef.current = pc;

			pc.addTransceiver("video", { direction: "recvonly" });
			pc.addTransceiver("audio", { direction: "recvonly" });

			pc.ontrack = (e) => {
				const el = videoRef.current;
				if (!el) return;
				if (e.streams && e.streams[0]) {
					el.srcObject = e.streams[0];
					el.onloadedmetadata = () => {
						if (el.videoWidth && el.videoHeight)
							onVideoDims?.(el.videoWidth, el.videoHeight);
					};
					el.play().catch(() => {});
				}
			};

			pc.onconnectionstatechange = () => {
				if (pc.connectionState === "connected") setStatus("playing");
				else if (
					pc.connectionState === "failed" ||
					pc.connectionState === "disconnected"
				) {
					handleError(new Error(`Peer connection ${pc.connectionState}`));
				}
			};

			try {
				const offer = await pc.createOffer();
				await pc.setLocalDescription(offer);
				await waitForIceComplete(pc);

				fetchAbortRef.current = new AbortController();
				const resp = await fetch(whepUrl, {
					method: "POST",
					headers: { "Content-Type": "application/sdp" },
					body: offer.sdp ?? "",
					signal: fetchAbortRef.current.signal,
				});

				if (!resp.ok) {
					const txt = await resp.text().catch(() => "");
					throw new Error(`WHEP POST ${resp.status}${txt ? `: ${txt}` : ""}`);
				}

				const location = resp.headers.get("Location");
				resourceUrlRef.current = location ?? null;

				const answerSdp = await resp.text();
				await pc.setRemoteDescription({ type: "answer", sdp: answerSdp });
			} catch (err) {
				handleError(err as Error);
			}
		};

		const handleError = (err: Error) => {
			if (cancelled) return;
			console.error(err);
			setStatus("error");
			onPlayError?.(err);
			if (autoRetry) setTimeout(() => !cancelled && setup(), retryDelayMs);
		};

		const cleanup = () => {
			fetchAbortRef.current?.abort();
			fetchAbortRef.current = null;
			if (resourceUrlRef.current) {
				fetch(resourceUrlRef.current, { method: "DELETE" }).catch(() => {});
			}
			resourceUrlRef.current = null;
			try {
				pcRef.current?.getSenders().forEach((s) => s.track?.stop());
				pcRef.current?.getReceivers().forEach((r) => r.track?.stop());
				pcRef.current?.close();
			} catch (err) {
				console.error(err);
			}
			pcRef.current = null;
		};

		setup();
		return () => {
			cancelled = true;
			cleanup();
		};
	}, [whepUrl, autoRetry, retryDelayMs, onPlayError, stunServers, onVideoDims]);

	return (
		<video
			ref={videoRef}
			autoPlay
			muted
			playsInline
			controls
			style={{
				width: "100%",
				height: "100%",
				objectFit: "contain",
				background: "black",
				display: "block",
			}}
			title={status}
		/>
	);
}

import { useEffect, useRef } from "react";

interface WebRTCProps {
	whepUrl: string;
}

export default function WebRTCPlayer({ whepUrl }: WebRTCProps) {
	const videoRef = useRef<HTMLVideoElement>(null);

	useEffect(() => {
		let peerConn: RTCPeerConnection | null = null;

		const goLive = async () => {
			peerConn = new RTCPeerConnection({
				iceServers: [
					{
						urls: ["stun:stun.l.google.com:19302"],
					},
				],
			});
			console.log("creating peer connection...");
			peerConn.addTransceiver("video", { direction: "recvonly" });
			peerConn.addTransceiver("audio", { direction: "recvonly" });

			peerConn.ontrack = (e) => {
				if (!videoRef.current) {
					console.warn("perrConn.ontrack(): videoRef.current is null");
					return;
				}

				videoRef.current.srcObject = e.streams[0];
				console.log(e.streams);
			};

			const offer = await peerConn.createOffer();
			await peerConn.setLocalDescription(offer);

			const req = await fetch(whepUrl, {
				method: "POST",
				headers: { "Content-Type": "application/sdp" },
				body: offer.sdp,
			});

			if (!req.ok) {
				const txt = await req.text();
				throw new Error(`WHEP POST ${req.status}: ${txt}`);
			}

			const answer = await req.text();
			await peerConn.setRemoteDescription({ type: "answer", sdp: answer });
		};
		goLive();

		return () => peerConn?.close();
	}, [whepUrl]);

	const enableAudio = async () => {
		const elem = videoRef.current;
		if (!elem) {
			console.warn("enableAudio(): videoRef.current is null");
			return;
		}
		elem.muted = !elem.muted;
		elem.volume = elem.muted ? 0.0 : 1.0;
		try {
			await elem.play();
		} catch (e) {
			console.error(e);
		}
	};

	return (
		<div>
			<video
				ref={videoRef}
				autoPlay
				muted
				playsInline
				controls
				style={{ width: 640, height: 640, borderRadius: "0.5%" }}
			></video>
			<button onClick={enableAudio}>Enable sound</button>
		</div>
	);
}

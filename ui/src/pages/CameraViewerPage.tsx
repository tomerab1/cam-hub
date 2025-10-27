import { useCallback, useEffect, useMemo, useState } from "react";
import { useParams } from "react-router-dom";
import {
	Box,
	Container,
	Typography,
	CardContent,
	Stack,
	Tooltip,
	IconButton,
	Skeleton,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";

import GlowCard from "../components/GlowCard";
import WebRTCPlayer from "../components/WebRTCPlayer";
import CameraDetailsCard from "../components/CameraDetailsCard";
import { useCameras } from "../providers/CamerasProvider";
import PtzControls from "../components/PtzControls";
import { NavigateBeforeOutlined } from "@mui/icons-material";
import CameraEventLog from "../components/CameraEventLog";

export default function CameraViewerPage() {
	const { id: cameraUUID } = useParams<{ id: string }>();
	const { cameras, getStreamUrl, invalidateStreamUrl, deleteStream } =
		useCameras();

	const currentCamera = useMemo(
		() => cameras.find((c) => c.uuid === cameraUUID),
		[cameras, cameraUUID]
	);

	const [streamUrl, setStreamUrl] = useState<string | null>(null);
	const [err, setErr] = useState<string | null>(null);
	const [copiedField, setCopiedField] = useState<string | null>(null);
	const [ratio, setRatio] = useState(16 / 9);

	useEffect(() => {
		// Used to prevent state updates when the comp is unmounted.
		let mounted = true;
		setErr(null);
		setStreamUrl(null);

		if (!cameraUUID) {
			setErr("Missing camera id in route.");
			return;
		}
		(async () => {
			try {
				const url = await getStreamUrl(cameraUUID);
				if (mounted) setStreamUrl(url + "/whep");
				if (mounted && !url) setErr("Stream URL unavailable.");
			} catch (e) {
				console.error(e);
				if (mounted) setErr(e instanceof Error ? e.message : "fetch failed");
			}
		})();

		return () => {
			mounted = false;
		};
	}, [cameraUUID, getStreamUrl]);

	const handleRetryStream = useCallback(() => {
		if (!cameraUUID) return;
		deleteStream(cameraUUID);
		invalidateStreamUrl(cameraUUID);
		setErr(null);
		setStreamUrl(null);
		getStreamUrl(cameraUUID).then((u) => setStreamUrl(u));
	}, [cameraUUID, invalidateStreamUrl, getStreamUrl, deleteStream]);

	const handleVideoDims = useCallback((w: number, h: number) => {
		if (h) setRatio(w / h);
	}, []);

	const onCopy = async (label: string, text?: string | null) => {
		if (!text) return;
		try {
			await navigator.clipboard.writeText(text);
			setCopiedField(label);
			setTimeout(() => setCopiedField(null), 1200);
		} catch (err) {
			console.error(err);
		}
	};

	return (
		<Box sx={{ minHeight: "100vh", bgcolor: "background.default", py: 4 }}>
			<Container maxWidth="xl">
				{/* header */}
				<Box sx={{ display: "flex", alignItems: "center", gap: 1, mb: 2 }}>
					<IconButton size="small" onClick={() => history.back()}>
						<NavigateBeforeOutlined />
					</IconButton>
					<Typography variant="h5" sx={{ flexGrow: 1 }}>
						{currentCamera?.camera_name ?? "Camera Viewer"}
					</Typography>
				</Box>
				<Box
					sx={{
						display: "grid",
						gridTemplateColumns: { xs: "1fr", md: "1fr 360px" },
						gap: 2,
						alignItems: "start",
					}}
				>
					{/* LEFT: player card */}
					<GlowCard
						sx={{
							gridColumn: "1",
							height: "100%",
						}}
					>
						<CardContent sx={{ p: { xs: 2, md: 2.5 } }}>
							<Box
								sx={{
									position: "relative",
									width: "100%",
									pt: `${(1 / ratio) * 70}%`,
									borderRadius: 2,
									overflow: "hidden",
									bgcolor: "black",
								}}
							>
								<Box sx={{ position: "absolute", inset: 0 }}>
									{!streamUrl ? (
										<Skeleton
											variant="rectangular"
											width="100%"
											height="100%"
										/>
									) : (
										<WebRTCPlayer
											whepUrl={streamUrl}
											onPlayError={handleRetryStream}
											onVideoDims={handleVideoDims}
										/>
									)}
								</Box>
							</Box>

							{err && (
								<Stack
									direction="row"
									alignItems="center"
									justifyContent="space-between"
									sx={{ mt: 1.5 }}
								>
									<Typography variant="body1" color="error" fontWeight={600}>
										{err}
									</Typography>
									<Tooltip title="Retry stream">
										<IconButton onClick={handleRetryStream} size="small">
											<RefreshIcon />
										</IconButton>
									</Tooltip>
								</Stack>
							)}
						</CardContent>
					</GlowCard>

					{/* RIGHT: details card & PTZ */}
					<Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
						<CameraDetailsCard
							camera={currentCamera}
							streamUrl={streamUrl ?? undefined}
							copiedField={copiedField}
							onCopy={onCopy}
							onRetry={handleRetryStream}
						/>
						<GlowCard sx={{ mt: 2 }}>
							<CardContent>
								<PtzControls uuid={cameraUUID!} />
							</CardContent>
						</GlowCard>
					</Box>
					{/* Event log - bottom left */}
					<Box sx={{ display: "flex", flexDirection: "column", gap: 2 }}>
						<CameraEventLog cameraUUID={cameraUUID} />
					</Box>
				</Box>
			</Container>
		</Box>
	);
}

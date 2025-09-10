import { Box, Button, Typography } from "@mui/material";
import { StatCard } from "../components/StatusCard";
import {
	CameraAltRounded,
	SensorsRounded,
	WarningAmberRounded,
} from "@mui/icons-material";
import WifiFindIcon from "@mui/icons-material/WifiFind";
import { CameraCard } from "../components/CameraCard";
import { useNavigate } from "react-router-dom";
import { useCameras } from "../providers/CamerasProvider";

export default function HomePage() {
	const navigate = useNavigate();
	const { cameras, loading, error } = useCameras();

	return (
		<Box
			sx={{
				display: "flex",
				flexDirection: "column",
				minHeight: "100vh",
				bgcolor: "background.default",
				alignItems: "center",
				p: 4,
				gap: 3,
			}}
		>
			<Box
				sx={{
					maxWidth: "80vw",
					width: "100%",
				}}
			>
				<Box
					sx={{
						display: "flex",
						flexDirection: "column",
						alignItems: "start",
						mb: "1.5rem",
					}}
				>
					<Box
						sx={{
							width: "100%",
							display: "flex",
							alignItems: "center",
							justifyContent: "space-between",
						}}
					>
						<Typography
							variant="h4"
							fontWeight="bold"
							sx={{ color: "oklch(97% 0.001 106.424)" }}
						>
							Security Dashboard
						</Typography>
						<Button
							sx={{
								color: "oklch(97% 0.001 106.424)",
								"& .MuiButton-startIcon > *:nth-of-type(1)": { fontSize: 25 },
							}}
							startIcon={<WifiFindIcon />}
							onClick={() => navigate("/discovery")}
						></Button>
					</Box>
					<Typography sx={{ color: "oklch(70.9% 0.01 56.259)" }}>
						Monitor and manage your home security cameras
					</Typography>
				</Box>
				{/* Status Cards */}
				<Box
					sx={{
						display: "grid",
						gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
						gap: 2.5,
					}}
				>
					<StatCard
						title="Total Cameras"
						value={8}
						icon={CameraAltRounded}
						desc="Configured cameras"
					/>
					<StatCard
						title="Online"
						value={7}
						valueColor="oklch(87.1% 0.15 154.449)"
						icon={SensorsRounded}
						desc="Active connections"
					/>
					<StatCard
						title="Alerts (24h)"
						value={3}
						valueColor="oklch(83.7% 0.128 66.29)"
						icon={WarningAmberRounded}
						desc="Alert pending"
					/>
				</Box>
				<Box
					sx={{
						display: "flex",
						alignItems: "center",
						marginTop: "4rem",
						justifyContent: "space-between",
						marginBottom: "1.5rem",
					}}
				>
					<Typography
						variant="h6"
						fontWeight="bold"
						sx={{ color: "oklch(97% 0.001 106.424)" }}
					>
						Your Cameras
					</Typography>
					<Typography
						fontWeight={400}
						sx={{
							display: "inline-flex",
							color: "oklch(97% 0.001 106.424)",
							border: "0.5px solid oklch(60.9% 0.01 56.259)",
							backgroundColor: "oklch(40.4% 0.01 67.558)",
							borderRadius: "9999px",
							paddingX: "0.625rem",
							paddingY: "0.125rem",
						}}
					>
						{cameras ? cameras.length : 0}{" "}
						{cameras && cameras.length == 1 ? "camera" : "cameras"}
					</Typography>
				</Box>
				<Box
					sx={{
						display: "grid",
						gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
						gap: 2.5,
					}}
				>
					{loading && (
						<Box sx={{ marginX: "auto", marginTop: "1.5rem" }}>
							<Typography
								variant="h6"
								fontWeight="bold"
								sx={{ color: "oklch(82% 0.001 106.424)" }}
							>
								Loading…
							</Typography>
						</Box>
					)}
					{error && (
						<Box sx={{ marginX: "auto", marginTop: "1.5rem" }}>
							<Typography
								variant="h6"
								fontWeight="bold"
								sx={{ color: "oklch(82% 0.001 106.424)" }}
							>
								Error: {error}
							</Typography>
						</Box>
					)}
					{!loading &&
						!error &&
						cameras?.map((cam) => (
							<CameraCard
								key={cam.uuid}
								id={cam.uuid}
								imgUri="/assets/backyard_camera.jpg"
								location={cam.camera_name || cam.uuid}
								status={cam.is_paired ? "online" : "offline"}
							/>
						))}
				</Box>
			</Box>
		</Box>
	);
}

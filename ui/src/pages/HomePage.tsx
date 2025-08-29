import { Box, Typography } from "@mui/material";
import { StatCard } from "../components/StatusCard";
import {
	CameraAltRounded,
	SensorsRounded,
	WarningAmberRounded,
} from "@mui/icons-material";
import { CameraCard } from "../components/CameraCard";

export default function HomePage() {
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
					<Typography
						variant="h4"
						fontWeight="bold"
						sx={{ color: "oklch(97% 0.001 106.424)" }}
					>
						Security Dashboard
					</Typography>
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
						4 cameras
					</Typography>
				</Box>
				<Box
					sx={{
						display: "grid",
						gridTemplateColumns: "repeat(auto-fit, minmax(260px, 1fr))",
						gap: 2.5,
					}}
				>
					<CameraCard
						imgUri="src/assets/backyard_camera.jpg"
						location="Backyard camera"
						status="offline"
					/>
					<CameraCard
						imgUri="src/assets/backyard_camera.jpg"
						location="Backyard camera"
						status="offline"
					/>
					<CameraCard
						imgUri="src/assets/backyard_camera.jpg"
						location="Backyard camera"
						status="offline"
					/>
					<CameraCard
						imgUri="src/assets/backyard_camera.jpg"
						location="Backyard camera"
						status="offline"
					/>
				</Box>
			</Box>
		</Box>
	);
}

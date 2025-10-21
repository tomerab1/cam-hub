import { Box, Card, Typography, alpha, CardMedia, Button } from "@mui/material";
import { useNavigate } from "react-router-dom";
import ClearIcon from "@mui/icons-material/Clear";
import { useState } from "react";

type Status = "online" | "offline";

type CameraCardProps = {
	id: string;
	imgUri?: string;
	location: string;
	status: Status;
};

export function CameraCard({ id, status, imgUri, location }: CameraCardProps) {
	const navigate = useNavigate();
	const image = imgUri ?? "/assets/default_camera.jpg";
	const [err, setError] = useState<string | null>(null);

	const unpairCamera = async () => {
		try {
			setError(null);
			const res = await fetch(
				`http://localhost:5555/api/v1/cameras/${id}/pair`,
				{ method: "DELETE" }
			);
			if (!res.ok) {
				throw new Error(`HTTP ${res.status}`);
			}
			navigate(0);
		} catch (e) {
			console.error(e);
			if (e instanceof Error) {
				setError(e.message ?? "fetch failed");
			} else {
				setError("Unknown error");
			}
		}
	};

	return (
		<Card
			elevation={1}
			sx={{
				position: "relative",
				borderRadius: 3,
				minWidth: 60,
				maxWidth: 300,
				overflow: "hidden",
				border: `1.75px solid ${alpha("#ffffff", 0.08)}`,
				boxShadow: `0 10px 30px ${alpha("#000", 0.25)}`,
				transition: "transform 0.15s ease, box-shadow 0.15s ease",
				"&:hover": {
					transform: "translateY(-1px)",
					boxShadow: `0 16px 40px ${alpha("#000", 0.35)}`,
				},
			}}
		>
			<Button
				variant="text"
				size="large"
				sx={{
					backgroundColor: "transparent",
					color: "black",
					border: "none",
					position: "absolute",
					padding: 0,
					marginTop: "5px",
					zIndex: 1,
					right: 0,
				}}
				endIcon={<ClearIcon />}
				onClick={unpairCamera}
			></Button>
			<CardMedia
				component="img"
				image={image}
				sx={{ height: 300, mb: "0.5rem" }}
			/>

			<Box sx={{ px: "1rem", py: "0.625rem" }}>
				<Box
					sx={{
						display: "flex",
						alignItems: "center",
						justifyContent: "space-between",
						mb: "1rem",
					}}
				>
					<Typography
						fontWeight={500}
						sx={{
							letterSpacing: 1,
							fontSize: "1.3rem",
							color: "oklch(97% 0.001 106.424)",
						}}
					>
						{location}
					</Typography>
					<Typography
						sx={{
							display: "inline-flex",
							color: "oklch(87% 0.001 106.424)",
							border: "0.5px solid oklch(60.9% 0.01 56.259)",
							backgroundColor:
								status === "online"
									? "oklch(67.1% 0.15 154.449)"
									: "oklch(63.7% 0.237 25.331)",
							borderRadius: "9999px",
							px: "0.625rem",
							py: "0.125rem",
						}}
					>
						{status}
					</Typography>
				</Box>

				<Button
					sx={{ color: "oklch(87% 0.001 106.424)" }}
					onClick={() => navigate(`/camera/${id}`)}
				>
					View camera
				</Button>
			</Box>
		</Card>
	);
}

import { Box, Button, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import type { DiscoveryDto } from "../contracts/DiscoveryDto";
import CameraDiscoveryTable from "../components/CameraDiscoveryTable";
import { NavigateBeforeOutlined } from "@mui/icons-material";
import { useNavigate } from "react-router-dom";

export default function CameraDiscoveryPage() {
	const [matches, setMatches] = useState<DiscoveryDto>({ matches: [] });
	const [isLoading, setLoading] = useState(false);
	const navigate = useNavigate();

	const discoverCameras = async () => {
		try {
			setLoading(true);
			const resp = await fetch("http://localhost:5555/api/v1/cameras");
			const matches: DiscoveryDto = await resp.json();
			setMatches(matches);
			console.log(matches);
		} catch (err) {
			console.error(err);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		discoverCameras();
	}, []);

	const onGoBack = () => {
		navigate("/");
	};

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
				<Box sx={{ display: "flex", alignItems: "center", mb: "1.5rem" }}>
					<Button
						aria-label="Go back to home"
						title="Go back to home"
						onClick={onGoBack}
						variant="text"
						sx={{ color: "black" }}
					>
						<NavigateBeforeOutlined />
					</Button>
					<Typography
						variant="h4"
						fontWeight="bold"
						sx={{ color: "oklch(97% 0.001 106.424)", marginLeft: "1rem" }}
					>
						Camera Discovery
					</Typography>
				</Box>

				<Box sx={{ display: "flex", justifyContent: "center" }}>
					<CameraDiscoveryTable
						discoverCameras={discoverCameras}
						isLoading={isLoading}
						matches={matches}
					/>
				</Box>
			</Box>
		</Box>
	);
}

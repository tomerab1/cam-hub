import { Box, Container, IconButton, Typography } from "@mui/material";
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
			const resp = await fetch("http://localhost:5555/api/v1/cameras/discovery");
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

	return (
		<Box sx={{ minHeight: "100vh", bgcolor: "background.default", py: 4 }}>
			<Container maxWidth="md">
				<Box sx={{ display: "flex", alignItems: "center", mb: 2 }}>
					<IconButton
						aria-label="Go back"
						onClick={() => navigate("/")}
						size="large"
						sx={{ color: "text.secondary", mr: 1 }}
					>
						<NavigateBeforeOutlined />
					</IconButton>
					<Typography
						variant="h4"
						fontWeight={800}
						sx={{ color: "text.primary" }}
					>
						Camera Discovery
					</Typography>
				</Box>

				<CameraDiscoveryTable
					discoverCameras={discoverCameras}
					isLoading={isLoading}
					matches={matches}
				/>
			</Container>
		</Box>
	);
}

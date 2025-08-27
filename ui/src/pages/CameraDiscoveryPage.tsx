import { Box, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import type { DiscoveryDto } from "../contracts/DiscoveryDto";
import CameraDiscoveryTable from "../components/CameraDiscoveryTable";

export default function CameraDiscoveryPage() {
	const [matches, setMatches] = useState<DiscoveryDto>({ matches: [] });
	const [isLoading, setLoading] = useState(false);

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

	return (
		<Box
			sx={{
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				minHeight: "100vh",
				bgcolor: "background.default",
				p: 4,
				gap: 3,
			}}
		>
			<Typography
				variant="h4"
				fontWeight="bold"
				sx={{ color: "oklch(92.3% 0.003 48.717)" }}
			>
				Camera Discovery
			</Typography>

			<CameraDiscoveryTable
				discoverCameras={discoverCameras}
				isLoading={isLoading}
				matches={matches}
			/>
		</Box>
	);
}

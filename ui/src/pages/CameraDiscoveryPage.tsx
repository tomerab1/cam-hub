import {
	Box,
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	Paper,
	Typography,
	IconButton,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import { useEffect, useState } from "react";

interface DiscoveryMatch {
	uuid: string;
	addr: string;
}

interface DiscoveryRes {
	matches: DiscoveryMatch[];
}

export default function CameraDiscoveryPage() {
	const [matches, setMatches] = useState<DiscoveryRes>();
	const [loading, setLoading] = useState(false);
	const [deviceToPair, setDeviceToPair] = useState<string>("");

	const discoverCameras = async () => {
		try {
			setLoading(true);
			const resp = await fetch("http://localhost:5555/api/v1/cameras");
			const matches: DiscoveryRes = await resp.json();
			setMatches(matches);
		} catch (err) {
			console.error(err);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		console.log(`pairing device with uuid=${deviceToPair}`);
	}, [deviceToPair]);

	return (
		<Box
			sx={{
				display: "flex",
				flexDirection: "column",
				alignItems: "center",
				justifyContent: "center",
				minHeight: "100vh",
				bgcolor: "background.default",
				p: 4,
				gap: 3,
			}}
		>
			<Typography variant="h4" fontWeight="bold">
				Camera Discovery
			</Typography>

			<TableContainer
				component={Paper}
				sx={{ width: "80%", maxWidth: 800, boxShadow: 3, borderRadius: 2 }}
			>
				<Table>
					<TableHead sx={{ bgcolor: "grey.200" }}>
						<TableRow>
							<TableCell sx={{ fontWeight: "bold" }}>UUID</TableCell>
							<TableCell sx={{ fontWeight: "bold" }}>Address</TableCell>
							<TableCell align="right">
								<IconButton
									onClick={discoverCameras}
									disabled={loading}
									color="primary"
								>
									<RefreshIcon
										sx={{
											animation: loading ? "spin 1s linear infinite" : "none",
											"@keyframes spin": {
												from: { transform: "rotate(0deg)" },
												to: { transform: "rotate(360deg)" },
											},
											color: "#3f3f3f",
										}}
									/>
								</IconButton>
							</TableCell>
						</TableRow>
					</TableHead>
					<TableBody>
						{matches?.matches.map((match, idx) => (
							<TableRow
								key={idx}
								hover
								sx={{
									"&:nth-of-type(odd)": { bgcolor: "grey.50" },
									"&:hover": {
										cursor: "pointer",
									},
								}}
								onClick={() => {
									setDeviceToPair(match.uuid);
								}}
							>
								<TableCell>{match.uuid}</TableCell>
								<TableCell>{match.addr}</TableCell>
								<TableCell />
							</TableRow>
						))}
						{!matches?.matches?.length && (
							<TableRow>
								<TableCell colSpan={3} align="center" sx={{ py: 3 }}>
									No cameras found. Click the refresh icon to scan.
								</TableCell>
							</TableRow>
						)}
					</TableBody>
				</Table>
			</TableContainer>
		</Box>
	);
}

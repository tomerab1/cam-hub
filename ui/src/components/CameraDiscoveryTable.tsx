import {
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	Paper,
	IconButton,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import type { DiscoveryDto } from "../contracts/DiscoveryDto";
import { useState } from "react";
import CameraPairingDialog from "./CameraPairingDialog";

interface CameraDiscoveryTableProps {
	discoverCameras: () => void;
	isLoading: boolean;
	matches: DiscoveryDto;
}

export default function CameraDiscoveryTable({
	discoverCameras,
	isLoading,
	matches,
}: CameraDiscoveryTableProps) {
	const [open, setOpen] = useState<boolean>(false);
	const [uuid, setUUID] = useState<string>("");
	const [addr, setAddr] = useState<string>("");

	const handleClickOpen = (uuid: string, addr: string) => {
		setOpen(true);
		setUUID(uuid);
		setAddr(addr);
	};

	const handleClose = () => {
		setOpen(false);
	};

	return (
		<TableContainer
			component={Paper}
			sx={{ width: "80%", maxWidth: 800, boxShadow: 3, borderRadius: 2 }}
		>
			<Table>
				<TableHead component={Paper}>
					<TableRow>
						<TableCell
							sx={{
								fontWeight: "bold",
								borderBottom: "none",
								color: "oklch(92.3% 0.003 48.717)",
							}}
						>
							UUID
						</TableCell>
						<TableCell
							sx={{
								fontWeight: "bold",
								borderBottom: "none",
								color: "oklch(92.3% 0.003 48.717)",
							}}
						>
							Address
						</TableCell>
						<TableCell align="right" sx={{ borderBottom: "none" }}>
							<IconButton
								onClick={discoverCameras}
								disabled={isLoading}
								color="primary"
							>
								<RefreshIcon
									sx={{
										animation: isLoading ? "spin 1s linear infinite" : "none",
										"@keyframes spin": {
											from: { transform: "rotate(0deg)" },
											to: { transform: "rotate(360deg)" },
										},
										color: "oklch(92.3% 0.003 48.717)",
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
							component={Paper}
							sx={{
								"&:hover": {
									cursor: "pointer",
								},
							}}
							onClick={() => {
								handleClickOpen(match.uuid, match.addr);
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
					{open && (
						<CameraPairingDialog
							addr={addr}
							uuid={uuid}
							open={open}
							handleClose={handleClose}
						/>
					)}
				</TableBody>
			</Table>
		</TableContainer>
	);
}

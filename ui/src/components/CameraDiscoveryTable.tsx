import {
	Box,
	Paper,
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	IconButton,
	Tooltip,
	CircularProgress,
	Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import RadarRounded from "@mui/icons-material/RadarRounded";
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
	const [open, setOpen] = useState(false);
	const [uuid, setUUID] = useState("");
	const [addr, setAddr] = useState("");

	const handleClickOpen = (u: string, a: string) => {
		setOpen(true);
		setUUID(u);
		setAddr(a);
	};
	const handleClose = () => setOpen(false);

	return (
		<TableContainer
			component={Paper}
			elevation={6}
			sx={{
				width: "100%",
				mx: 0,
				borderRadius: 3,
				overflow: "hidden",
				backdropFilter: "blur(4px)",
			}}
		>
			<Table size="small" aria-label="discovered cameras">
				<TableHead>
					<TableRow>
						<TableCell
							sx={{
								fontWeight: 700,
								color: "text.secondary",
								borderBottom: (t) => `1px solid ${t.palette.divider}`,
							}}
						>
							UUID
						</TableCell>
						<TableCell
							sx={{
								fontWeight: 700,
								color: "text.secondary",
								borderBottom: (t) => `1px solid ${t.palette.divider}`,
							}}
						>
							Address
						</TableCell>
						<TableCell
							align="right"
							sx={{
								width: 1,
								borderBottom: (t) => `1px solid ${t.palette.divider}`,
							}}
						>
							<Tooltip title="Scan for cameras">
								<span>
									<IconButton
										onClick={discoverCameras}
										disabled={isLoading}
										size="small"
									>
										{isLoading ? (
											<CircularProgress size={18} />
										) : (
											<RefreshIcon sx={{ fontSize: 20 }} />
										)}
									</IconButton>
								</span>
							</Tooltip>
						</TableCell>
					</TableRow>
				</TableHead>

				<TableBody>
					{matches?.matches.map((m, idx) => (
						<TableRow
							key={idx}
							hover
							sx={{
								cursor: "pointer",
								"&:hover": { backgroundColor: (t) => t.palette.action.hover },
							}}
							onClick={() => handleClickOpen(m.uuid, m.addr)}
						>
							<TableCell
								sx={{
									fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
								}}
							>
								{m.uuid}
							</TableCell>
							<TableCell>{m.addr}</TableCell>
							<TableCell />
						</TableRow>
					))}

					{!matches?.matches?.length && (
						<TableRow>
							<TableCell colSpan={3} align="center" sx={{ py: 5 }}>
								<Box
									sx={{
										display: "flex",
										alignItems: "center",
										gap: 1,
										justifyContent: "center",
									}}
								>
									<RadarRounded sx={{ opacity: 0.7 }} />
									<Typography variant="body2" color="text.secondary">
										No cameras found. Click the refresh icon to scan.
									</Typography>
								</Box>
							</TableCell>
						</TableRow>
					)}
				</TableBody>
			</Table>

			{open && (
				<CameraPairingDialog
					addr={addr}
					uuid={uuid}
					open={open}
					handleClose={handleClose}
				/>
			)}
		</TableContainer>
	);
}

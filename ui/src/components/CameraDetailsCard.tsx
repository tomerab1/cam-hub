import * as React from "react";
import {
	CardContent,
	CardHeader,
	Stack,
	Typography,
	Divider,
	Chip,
	Tooltip,
	IconButton,
	List,
	ListItem,
	ListItemText,
} from "@mui/material";
import ContentCopyIcon from "@mui/icons-material/ContentCopy";
import RefreshIcon from "@mui/icons-material/Refresh";
import GlowCard from "../components/GlowCard";
import type { CameraDto } from "../contracts/CameraDto";

type Props = {
	camera?: CameraDto;
	streamUrl?: string | null;
	copiedField: string | null;
	onCopy: (label: string, value?: string | null) => void;
	onRetry: () => void;
};

export default function CameraDetailsCard({
	camera,
	streamUrl,
	copiedField,
	onCopy,
	onRetry,
}: Props) {
	const items = [
		{ label: "UUID", value: camera?.uuid, copyable: true },
		{ label: "Manufacturer", value: camera?.manufacturer },
		{ label: "Model", value: camera?.model },
		{ label: "Firmware", value: camera?.firmware_version },
		{ label: "Serial #", value: camera?.serial_number },
		{ label: "Hardware ID", value: camera?.hardware_id },
		{
			label: "WHEP Endpoint",
			value: streamUrl ?? undefined,
			copyable: true,
			mono: true,
		},
	];

	return (
		<GlowCard
			sx={{ height: "fit-content", display: "flex", flexDirection: "column" }}
		>
			<CardHeader
				title="Camera Details"
				action={
					<Tooltip title="Retry stream">
						<IconButton size="small" onClick={onRetry}>
							<RefreshIcon />
						</IconButton>
					</Tooltip>
				}
				sx={{ pb: 0, "& .MuiCardHeader-title": { fontWeight: 600 } }}
			/>
			<CardContent sx={{ pt: 1.5, pb: 2 }}>
				<Stack spacing={1.25}>
					<Stack
						direction="row"
						spacing={1}
						alignItems="center"
						flexWrap="wrap"
					>
						{camera?.addr && (
							<Chip label={camera.addr} size="small" variant="outlined" />
						)}
					</Stack>

					<Divider />

					<List dense disablePadding>
						{items.map((it) => (
							<ListItem
								key={it.label}
								disableGutters
								secondaryAction={
									it.copyable && it.value ? (
										<Tooltip
											title={copiedField === it.label ? "Copied!" : "Copy"}
										>
											<IconButton
												edge="end"
												size="small"
												onClick={() => onCopy(it.label, it.value)}
											>
												<ContentCopyIcon fontSize="small" />
											</IconButton>
										</Tooltip>
									) : undefined
								}
								sx={{ py: 0.25 }}
							>
								<ListItemText
									primary={
										<Typography variant="subtitle2" sx={{ opacity: 0.7 }}>
											{it.label}
										</Typography>
									}
									secondary={
										<Typography
											variant="body2"
											sx={{
												mt: 0.25,
												maxWidth: "100%",
												whiteSpace: "nowrap",
												overflow: "hidden",
												textOverflow: "ellipsis",
												fontFamily: it.mono
													? "ui-monospace, SFMono-Regular, Menlo, monospace"
													: undefined,
											}}
											title={(it.value as string) || ""}
										>
											{(it.value as string) || "—"}
										</Typography>
									}
									secondaryTypographyProps={{ component: "div" }}
								/>
							</ListItem>
						))}
					</List>
				</Stack>
			</CardContent>
		</GlowCard>
	);
}

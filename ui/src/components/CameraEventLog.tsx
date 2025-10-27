import { alpha, CardContent, List, Typography, useTheme } from "@mui/material";
import { useEffect, useState } from "react";
import { type CameraEventDto } from "../contracts/CameraEventDto";
import CircleIcon from "@mui/icons-material/Circle";
import GlowCard from "./GlowCard";

interface CameraEventLogProps {
	cameraUUID?: string;
}

function getCircleColor(score: number): string {
	if (0.5 <= score && score <= 0.6) return "oklch(85.2% 0.199 91.936)";
	if (0.6 <= score && score <= 0.7) return "oklch(75% 0.183 55.934)";
	return "oklch(70.4% 0.191 22.216)";
}

function getDescription(score: number): string {
	if (0.5 <= score && score <= 0.6) return "Low confidence (possible person)";
	if (0.6 <= score && score <= 0.7) return "Medium confidence (likely person)";
	return "High confidence (confirmed person)";
}

export default function CameraEventLog({ cameraUUID }: CameraEventLogProps) {
	const theme = useTheme();
	const glow = alpha(theme.palette.primary.main, 0.25);
	const [events, setEvents] = useState<CameraEventDto[]>([]);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		const es = new EventSource(
			`http://localhost:5555/api/v1/events/recordings/${cameraUUID}`
		);
		es.onmessage = (e) => {
			try {
				setError(null);
				const evt: CameraEventDto = JSON.parse(e.data);
				setEvents((prev) => [evt, ...prev].slice(0, 4));
			} catch (err) {
				console.error(err);
				if (err instanceof Error) setError(err.message);
				else setError("Unknown error");
			}
		};
		return () => es.close();
	}, [cameraUUID, events]);

	return (
		<List>
			{events.map((ev) => (
				<GlowCard
					key={ev.id}
					sx={{
						marginTop: "0.5rem",
					}}
				>
					<CardContent
						sx={{
							display: "flex",
							alignItems: "center",
							paddingInline: "1rem",
							gap: 1.5,
						}}
					>
						<CircleIcon
							fontSize="medium"
							htmlColor={getCircleColor(ev.score)}
						/>
						<Typography
							variant="h6"
							sx={{ letterSpacing: 1, color: { glow }, opacity: 0.8 }}
						>
							{getDescription(ev.score)}
						</Typography>
					</CardContent>
				</GlowCard>
			))}
		</List>
	);
}

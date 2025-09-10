import {
	KeyboardArrowDownOutlined,
	KeyboardArrowLeftOutlined,
	KeyboardArrowRightOutlined,
	KeyboardArrowUpOutlined,
} from "@mui/icons-material";
import { Box, IconButton } from "@mui/material";
import { useCallback, useEffect, useRef } from "react";
import { type MoveCameraDto } from "../contracts/MoveCameraDto";

interface PtzControlsProps {
	uuid: string;
}

const NUDGE = 0.12;
const REPEAT_MS = 150;
const COOLDOWN_MS = 120;

export default function PtzControls({ uuid }: PtzControlsProps) {
	const repeatTimer = useRef<number | null>(null);
	const lastSentAt = useRef<number>(0);

	const sendMove = useCallback(
		async (dto: MoveCameraDto) => {
			const now = Date.now();
			if (now - lastSentAt.current < COOLDOWN_MS) return;
			lastSentAt.current = now;

			try {
				await fetch(`http://localhost:5555/api/v1/cameras/${uuid}/ptz/move`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(dto),
				});
			} catch (err) {
				console.error(err);
			}
		},
		[uuid]
	);

	const nudge = useCallback(
		(dx = 0, dy = 0) => {
			const dto: MoveCameraDto = { translation: { x: dx, y: dy } };
			void sendMove(dto);
		},
		[sendMove]
	);

	const startHold = useCallback(
		(dx = 0, dy = 0) => {
			nudge(dx, dy);
			stopHold();
			repeatTimer.current = window.setInterval(() => nudge(dx, dy), REPEAT_MS);
		},
		[nudge]
	);

	const stopHold = useCallback(() => {
		if (repeatTimer.current) {
			clearInterval(repeatTimer.current);
			repeatTimer.current = null;
		}
	}, []);

	useEffect(() => {
		const onBlur = () => stopHold();
		window.addEventListener("blur", onBlur);
		document.addEventListener("visibilitychange", onBlur);
		return () => {
			window.removeEventListener("blur", onBlur);
			document.removeEventListener("visibilitychange", onBlur);
			stopHold();
		};
	}, [stopHold]);

	return (
		<Box sx={{ height: "100%", display: "flex", alignItems: "center" }}>
			<div
				style={{
					display: "grid",
					gridTemplateColumns: "repeat(3, 40px)",
					gridTemplateRows: "repeat(3, 40px)",
					gridTemplateAreas: `". up ."
                              "left . right"
                              ". down ."`,
					placeItems: "center",
					gap: 6,
				}}
			>
				<IconButton
					sx={{ gridArea: "up" }}
					onClick={() => nudge(0, NUDGE)}
					onPointerDown={() => startHold(0, NUDGE)}
					onPointerUp={stopHold}
					onPointerLeave={stopHold}
				>
					<KeyboardArrowUpOutlined />
				</IconButton>

				<IconButton
					sx={{ gridArea: "right" }}
					onClick={() => nudge(-NUDGE, 0)}
					onPointerDown={() => startHold(-NUDGE, 0)}
					onPointerUp={stopHold}
					onPointerLeave={stopHold}
				>
					<KeyboardArrowRightOutlined />
				</IconButton>

				<IconButton
					sx={{ gridArea: "left" }}
					onClick={() => nudge(NUDGE, 0)}
					onPointerDown={() => startHold(NUDGE, 0)}
					onPointerUp={stopHold}
					onPointerLeave={stopHold}
				>
					<KeyboardArrowLeftOutlined />
				</IconButton>

				<IconButton
					sx={{ gridArea: "down" }}
					onClick={() => nudge(0, -NUDGE)}
					onPointerDown={() => startHold(0, -NUDGE)}
					onPointerUp={stopHold}
					onPointerLeave={stopHold}
				>
					<KeyboardArrowDownOutlined />
				</IconButton>
			</div>
		</Box>
	);
}

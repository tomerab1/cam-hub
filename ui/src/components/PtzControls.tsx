import {
	AddOutlined,
	KeyboardArrowDownOutlined,
	KeyboardArrowLeftOutlined,
	KeyboardArrowRightOutlined,
	KeyboardArrowUpOutlined,
	RemoveOutlined,
} from "@mui/icons-material";
import { Box, IconButton, Tooltip } from "@mui/material";
import { useCallback, useEffect, useRef } from "react";
import { type MoveCameraDto } from "../contracts/MoveCameraDto";

interface PtzControlsProps {
	uuid: string;
}

type ZoomAction = "in" | "out";

const NUDGE = 0.12;
const COOLDOWN_MS = 200;

export default function PtzControls({ uuid }: PtzControlsProps) {
	const lastSentAt = useRef(0);

	const sendMove = useCallback(
		async function (dx = 0, dy = 0) {
			const now = Date.now();
			if (now - lastSentAt.current < COOLDOWN_MS) return;
			lastSentAt.current = now;

			const dto: MoveCameraDto = { translation: { x: dx, y: dy }, zoom: null };
			try {
				await fetch(`http://localhost:5555/api/v1/cameras/${uuid}/ptz/move`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(dto),
					keepalive: true,
				});
			} catch (err) {
				console.error(err);
			}
		},
		[uuid]
	);

	const sendZoom = useCallback(
		async function (factor: number) {
			const now = Date.now();
			if (now - lastSentAt.current < COOLDOWN_MS) return;
			lastSentAt.current = now;
			const dto: MoveCameraDto = {
				translation: null,
				zoom: factor,
			};
			try {
				await fetch(`http://localhost:5555/api/v1/cameras/${uuid}/ptz/move`, {
					method: "POST",
					headers: { "Content-Type": "application/json" },
					body: JSON.stringify(dto),
					keepalive: true,
				});
			} catch (err) {
				console.error(err);
			}
		},
		[uuid]
	);

	useEffect(() => {
		function onKeyDown(e: KeyboardEvent) {
			switch (e.key) {
				case "ArrowUp":
					sendMove(0, NUDGE);
					break;
				case "ArrowDown":
					sendMove(0, -NUDGE);
					break;
				case "ArrowLeft":
					sendMove(NUDGE, 0);
					break;
				case "ArrowRight":
					sendMove(-NUDGE, 0);
					break;
				default:
					return;
			}
			e.preventDefault();
		}
		window.addEventListener("keydown", onKeyDown);
		return () => window.removeEventListener("keydown", onKeyDown);
	}, [sendMove]);

	const mkCtrlBtn = (
		area: string,
		dx: number,
		dy: number,
		Icon: React.ElementType,
		label: string
	) => (
		<Tooltip title={label}>
			<IconButton sx={{ gridArea: area }} onClick={() => sendMove(dx, dy)}>
				<Icon />
			</IconButton>
		</Tooltip>
	);

	const mkZoomBtn = (
		action: ZoomAction,
		Icon: React.ElementType,
		factor: number = 0.001
	) => {
		return (
			<Tooltip title={action}>
				<IconButton
					onClick={() => sendZoom(action === "in" ? factor : -factor)}
				>
					<Icon />
				</IconButton>
			</Tooltip>
		);
	};

	return (
		<Box
			sx={{
				height: "100%",
				display: "flex",
				alignItems: "center",
				justifyContent: "space-between",
			}}
		>
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
				{mkCtrlBtn("up", 0, NUDGE, KeyboardArrowUpOutlined, "Up")}
				{mkCtrlBtn("right", -NUDGE, 0, KeyboardArrowRightOutlined, "Right")}
				{mkCtrlBtn("left", NUDGE, 0, KeyboardArrowLeftOutlined, "Left")}
				{mkCtrlBtn("down", 0, -NUDGE, KeyboardArrowDownOutlined, "Down")}
			</div>
			<Box sx={{ marginRight: "1rem" }}>
				{mkZoomBtn("in", AddOutlined)}
				{mkZoomBtn("out", RemoveOutlined)}
			</Box>
		</Box>
	);
}

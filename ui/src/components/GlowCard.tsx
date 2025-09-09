import { Card, type CardProps } from "@mui/material";
import { alpha, useTheme } from "@mui/material/styles";

export default function GlowCard({ sx, ...props }: CardProps) {
	const theme = useTheme();
	const border = alpha(theme.palette.common.white, 0.08);
	const glow = alpha(theme.palette.primary.main, 0.25);
	const shadow = alpha("#000", 0.35);

	return (
		<Card
			elevation={0}
			{...props}
			sx={{
				position: "relative",
				overflow: "hidden",
				borderRadius: 3,
				bgcolor: "background.paper",
				border: `1px solid ${border}`,
				boxShadow: `0 10px 30px ${shadow}`,
				"&::after": {
					content: '""',
					position: "absolute",
					inset: -2,
					borderRadius: 12,
					pointerEvents: "none",
					background: `radial-gradient(120% 80% at 0% 0%, ${glow}, transparent 55%)`,
					opacity: 0.45,
					mixBlendMode: "soft-light",
				},
				"&::before": {
					content: '""',
					position: "absolute",
					inset: 0,
					borderRadius: 12,
					pointerEvents: "none",
					boxShadow: `inset 0 1px 0 ${alpha("#fff", 0.06)}`,
				},
				...sx,
			}}
		/>
	);
}

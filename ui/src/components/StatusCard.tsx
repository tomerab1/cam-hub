import { Box, Card, Typography, SvgIcon, alpha } from "@mui/material";

type StatCardProps = {
	title: string;
	value: string | number;
	desc: string;
	valueColor?: string;
	icon?: React.ElementType;
};

export function StatCard({
	title,
	value,
	icon: Icon,
	desc,
	valueColor,
}: StatCardProps) {
	return (
		<Card
			elevation={1}
			sx={{
				position: "relative",
				borderRadius: 3,
				p: 2.5,
				minWidth: 260,
				overflow: "hidden",
				border: `1.75px solid ${alpha("#ffffff", 0.08)}`,

				boxShadow: `0 10px 30px ${alpha("#000", 0.25)}`,
				transition: "transform 0.15s ease, box-shadow 0.15s ease",
				"&:hover": {
					transform: "translateY(-1px)",
					boxShadow: `0 16px 40px ${alpha("#000", 0.35)}`,
				},
			}}
		>
			{/* Icon bubble */}
			{Icon && (
				<Box
					sx={{
						position: "absolute",
						right: 16,
						top: 16,
						width: 44,
						height: 44,
						borderRadius: "50%",
						display: "grid",
						placeItems: "center",
						backgroundColor: "oklch(44.4% 0.011 73.639 / 0.6)",
						outline: `1px solid ${alpha("#fff", 0.06)}`,
					}}
				>
					<SvgIcon component={Icon} sx={{ fontSize: 24, opacity: 0.85 }} />
				</Box>
			)}
			{/* Title */}
			<Typography
				variant="overline"
				sx={{ letterSpacing: 1, color: "oklch(82% 0.008 100)" }}
			>
				{title}
			</Typography>
			{/* Body */}
			<Typography
				variant="h4"
				sx={{
					mt: 0.5,
					fontWeight: 800,
					color: `${!valueColor ? " oklch(97% 0.001 106.424)" : valueColor}`,
				}}
			>
				{value}
			</Typography>
			{/* Footer */}
			<Typography
				variant="overline"
				sx={{ letterSpacing: 1, color: "oklch(82% 0.008 100)" }}
			>
				{desc}
			</Typography>
		</Card>
	);
}

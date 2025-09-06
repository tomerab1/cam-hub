import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { createTheme, styled, ThemeProvider } from "@mui/material";
import { MaterialDesignContent, SnackbarProvider } from "notistack";
import DiscoverySSE from "./components/SSEDiscovery.tsx";

const customTheme = createTheme({
	palette: {
		mode: "dark",
		background: {
			default: "oklch(44.4% 0.011 73.639)",
			paper: "oklch(44.4% 0.011 73.639)",
		},
	},
});

const StyledMaterialDesignContent = styled(MaterialDesignContent)(() => ({
	"&.notistack-MuiContent-success": {
		backgroundColor: "oklch(79.2% 0.209 151.711)",
	},
	"&.notistack-MuiContent-error": {
		backgroundColor: "oklch(63.7% 0.237 25.331)",
	},
}));

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<SnackbarProvider
			maxSnack={3}
			autoHideDuration={2000}
			Components={{
				success: StyledMaterialDesignContent,
				error: StyledMaterialDesignContent,
			}}
		>
			<ThemeProvider theme={customTheme}>
				<DiscoverySSE />
				<App />
			</ThemeProvider>
		</SnackbarProvider>
	</StrictMode>
);

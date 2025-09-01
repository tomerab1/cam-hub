import "@fontsource/roboto/300.css";
import "@fontsource/roboto/400.css";
import "@fontsource/roboto/500.css";
import "@fontsource/roboto/700.css";

import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { createTheme, ThemeProvider } from "@mui/material";

const customTheme = createTheme({
	palette: {
		mode: "dark",
		background: {
			default: "oklch(44.4% 0.011 73.639)",
			paper: "oklch(44.4% 0.011 73.639)",
		},
	},
});

createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<ThemeProvider theme={customTheme}>
			<App />
		</ThemeProvider>
	</StrictMode>
);

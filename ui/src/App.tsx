import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CameraDiscoveryPage from "./pages/CameraDiscoveryPage";
import HomePage from "./pages/HomePage";
import CameraViewerPage from "./pages/CameraViewerPage";

export default function App() {
	return (
		<Router>
			<Routes>
				<Route path="/" element={<HomePage />} />
				<Route path="/discovery" element={<CameraDiscoveryPage />} />
				<Route path="/camera/:cameraId" element={<CameraViewerPage />} />
			</Routes>
		</Router>
	);
}

import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import WebRTCPlayer from "./components/WebRTCPlayer";
import CameraDiscoveryPage from "./pages/CameraDiscoveryPage";
import HomePage from "./pages/HomePage";

export default function App() {
	const whepUrl = "http://localhost:8889/cam_01/whep";

	return (
		<Router>
			<Routes>
				<Route path="/" element={<HomePage />} />
				<Route path="/discovery" element={<CameraDiscoveryPage />} />
				<Route
					path="/player/:cameraId"
					element={<WebRTCPlayer whepUrl={whepUrl} />}
				/>
			</Routes>
		</Router>
	);
}

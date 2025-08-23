import WebRTCPlayer from "./components/player/WebRTCPlayer";
import CameraDiscoveryPage from "./pages/CameraDiscoveryPage";

export default function App() {
	const whepUrl = "http://localhost:8889/cam_01/whep";

	// return <WebRTCPlayer whepUrl={whepUrl} />;
	return <CameraDiscoveryPage />;
}

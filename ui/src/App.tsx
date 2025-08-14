import WebRTCPlayer from "./components/player/WebRTCPlayer";

export default function App() {
	const whepUrl = "http://localhost:8889/test_cam/whep";

	return <WebRTCPlayer whepUrl={whepUrl} />;
}

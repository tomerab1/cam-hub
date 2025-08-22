import WebRTCPlayer from "./components/player/WebRTCPlayer";

export default function App() {
	const whepUrl = "http://localhost:8889/cam_01/whep";

	return <WebRTCPlayer whepUrl={whepUrl} />;
}

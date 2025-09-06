import { useSnackbar } from "notistack";
import { useEffect } from "react";

export default function DiscoverySSE() {
	const { enqueueSnackbar } = useSnackbar();

	useEffect(() => {
		const sse = new EventSource(
			"http://localhost:5555/api/v1/events/discovery"
		);

		sse.onmessage = (e) => {
			try {
				const evt = JSON.parse(e.data);
				if (evt.type === "device_new") {
					enqueueSnackbar(`New device: ${evt.uuid} @ ${evt.addr}`, {
						variant: "success",
					});
				} else if (evt.type === "device_ip_changed") {
					enqueueSnackbar(`IP changed: ${evt.uuid} → ${evt.addr}`, {
						variant: "success",
					});
				} else if (evt.type === "device_offline") {
					enqueueSnackbar(`Device offline: ${evt.uuid}`, {
						variant: "warning",
					});
				}
			} catch (err) {
				console.error(err);
			}
		};

		return () => sse.close();
	}, [enqueueSnackbar]);

	return null;
}

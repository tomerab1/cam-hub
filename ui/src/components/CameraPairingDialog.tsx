import * as React from "react";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import type { PairCameraDto } from "../contracts/PairCameraDto";

interface CameraPairingDialogProps {
	uuid: string;
	addr: string;
	open: boolean;
	handleClose: () => void;
}

export default function CameraPairingDialog({
	uuid,
	addr,
	open,
	handleClose,
}: CameraPairingDialogProps) {
	const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
		event.preventDefault();
		const formData = new FormData(event.currentTarget);
		const formJson = Object.fromEntries(formData.entries());
		const pairReq: PairCameraDto = {
			...(formJson as Omit<PairCameraDto, "addr">),
			addr,
		};

		const sendPairingReq = async () => {
			try {
				console.log(pairReq);
				const resp = await fetch(
					`http://localhost:5555/api/v1/cameras/${uuid}/pair`,
					{
						method: "POST",
						mode: "cors",
						headers: { "Content-Type": "application/json" },
						body: JSON.stringify(pairReq),
					}
				);

				console.log(resp);
			} catch (err) {
				console.error(err);
			}
			handleClose();
		};

		sendPairingReq();
	};

	return (
		<React.Fragment>
			<Dialog open={open} onClose={handleClose}>
				<DialogTitle>Pair camera</DialogTitle>
				<DialogContent>
					<DialogContentText>
						Please enter a username, password and camera name.
					</DialogContentText>
					<form onSubmit={handleSubmit} id="subscription-form">
						<TextField
							autoFocus
							required
							margin="dense"
							name="username"
							label="Username"
							type="text"
							fullWidth
							variant="standard"
						/>
						<TextField
							autoFocus
							required
							margin="dense"
							name="password"
							label="Password"
							type="password"
							fullWidth
							variant="standard"
						/>
						<TextField
							autoFocus
							required
							margin="dense"
							name="camera_name"
							label="Camera Name"
							type="text"
							fullWidth
							variant="standard"
						/>
					</form>
				</DialogContent>
				<DialogActions>
					<Button onClick={handleClose}>Cancel</Button>
					<Button type="submit" form="subscription-form">
						Submit
					</Button>
				</DialogActions>
			</Dialog>
		</React.Fragment>
	);
}

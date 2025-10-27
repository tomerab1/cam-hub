export interface Evidence {
	conf: number;
	x_min: number;
	x_max: number;
	y_min: number;
	y_max: number;
}

export interface CameraEventDto {
	id: string;
	cam_id: string;
	bucket_name: string;
	vid_key: string;
	best_frame_key: string;
	evidence: Evidence;
	score: number;
	promoted_at: string;
	retention_days: number;
}

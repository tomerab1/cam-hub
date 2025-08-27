export interface DiscoveryMatch {
	uuid: string;
	addr: string;
}

export interface DiscoveryDto {
	matches: DiscoveryMatch[];
}

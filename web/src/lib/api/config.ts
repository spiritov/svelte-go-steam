import { PUBLIC_API_URL } from '$env/static/public';

interface JumpConfig {
	apiBaseUrl: string;
}

export default {
	apiBaseUrl: PUBLIC_API_URL
} as JumpConfig;

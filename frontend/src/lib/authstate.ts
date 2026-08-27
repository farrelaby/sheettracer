export interface OAuthState {
  connected: boolean;
  status: "disconnected" | "connecting" | "connected" | "expired";
  message?: string;
}

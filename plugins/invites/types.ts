export type Invite = {
  id: string;
  code: string;
  projectId: string;
  roomId: string;
  targetPlayerId: string;
  expiresAt: number;
  used: boolean;
};

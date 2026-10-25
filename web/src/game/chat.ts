import type { PlayerId, ServerMessage } from "@/net/protocol";

/** How long a chat balloon stays above a raccoon (spec, T23). */
export const BALLOON_MS = 5000;

/** RN4, mirrored from the server; the server still enforces it. */
export const MAX_MESSAGE_LENGTH = 120;

/**
 * Returns the text to send, or null when the server would drop it anyway:
 * blank, or longer than MAX_MESSAGE_LENGTH characters. Characters are code
 * points, like the server's rune count, so "🦝" counts as one.
 */
export function prepareMessage(text: string): string | null {
  if (text.trim() === "") return null;
  if ([...text].length > MAX_MESSAGE_LENGTH) return null;
  return text;
}

interface Balloon {
  text: string;
  at: number;
}

/** The current chat balloon of each player, fed as messages arrive. */
export class ChatBalloons {
  private readonly balloons = new Map<PlayerId, Balloon>();

  ingest(msg: ServerMessage, now: number): void {
    switch (msg.type) {
      case "chat.message":
        this.balloons.set(msg.data.from, { text: msg.data.text, at: now });
        break;
      case "player.left":
        this.balloons.delete(msg.data.id);
        break;
      case "room.joined":
        this.balloons.clear();
        break;
    }
  }

  /** The balloon text to show above a player at `now`, or null. */
  textFor(id: PlayerId, now: number): string | null {
    const b = this.balloons.get(id);
    if (!b) return null;
    if (now - b.at >= BALLOON_MS) {
      this.balloons.delete(id);
      return null;
    }
    return b.text;
  }
}

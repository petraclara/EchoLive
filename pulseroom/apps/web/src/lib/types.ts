export type Event = {
  id: string;
  organizer_id: string;
  title: string;
  slug: string;
  join_code: string;
  status: "draft" | "live" | "ended";
  created_at: string;
  updated_at: string;
};

export type Announcement = {
  id: string;
  event_id: string;
  body: string;
  type: "info" | "alert" | "link";
  link_url?: string;
  is_pinned: boolean;
  created_at: string;
};

export type Resource = {
  id: string;
  event_id: string;
  title: string;
  url?: string;
  kind: string;
  sort_order: number;
};

export type AgendaItem = {
  id: string;
  event_id: string;
  title: string;
  speaker?: string;
  sort_order: number;
};

export type PublicEvent = Event & {
  pinned?: Announcement;
  announcements: Announcement[];
  resources: Resource[];
  agenda: AgendaItem[];
  attendee_count: number;
};

export type WSMessage = {
  type: string;
  event_id: string;
  payload: unknown;
  ts: number;
};

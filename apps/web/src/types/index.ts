export interface Category {
  id: number;
  name: string;
  slug: string;
  color_code: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
}

export interface Event {
  id: string;
  title: string;
  description: string;
  lat: number; 
  lon: number; 
  start_time: string;
  max_participants: number;
  category_id: number;
  creator_id: string;
}
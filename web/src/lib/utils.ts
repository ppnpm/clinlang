import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// cn: the shadcn standard helper. Combines clsx (conditional classes)
// with tailwind-merge (dedupes conflicting utility classes).
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

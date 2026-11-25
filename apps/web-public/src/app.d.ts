declare global {
	namespace App {
		interface Error {
			message: string;
			code?: string;
		}
		interface Locals {
			user: { id: string; email: string; fullName: string; onboarded: boolean } | null;
		}
	}
}
export {};

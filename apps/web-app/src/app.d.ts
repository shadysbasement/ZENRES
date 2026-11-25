declare global {
	namespace App {
		interface Error {
			message: string;
			code?: string;
		}
		interface Locals {
			opsAdmin: { handle: string; fingerprint: string } | null;
		}
	}
}
export {};

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			session: Promise<Session | undefined>;
		}
		interface PageData {
			session: Session | undefined;
		}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    // interface Locals {}
    // interface PageData {}
    interface PageState {
      /** Set by the header's sign-out on its way to the login page, which opens on the signed-out line. */
      signedOut?: boolean;
    }
    // interface Platform {}
  }
}

export {};

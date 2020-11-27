declare module 'svelte-routing';

export class link {
  to: string;
  replace: boolean;
  state: object;
  getProps: () => object;
}

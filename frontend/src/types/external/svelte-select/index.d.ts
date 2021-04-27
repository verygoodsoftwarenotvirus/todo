declare module 'svelte-select';

declare interface SelectOption {
  value: string;
  label: string;
}

declare interface SelectedValue {
  detail: SelectOption;
}

declare interface SelectedValues {
  detail: SelectOption[];
}

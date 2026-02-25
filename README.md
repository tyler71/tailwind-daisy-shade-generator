# tailwind-daisy-shade-generator
Takes a `oklch`, `hex` or `hsl` value and generates tailwind daisyui color shades

```bash
Usage: tailwind-color-gen [--name primary] <color>
Examples:
  tailwind-color-gen --name brand "oklch(63.092% 0.20837 34.93)"
  tailwind-color-gen "#0ea5e9"
  tailwind-color-gen "hsl(200 80% 50%)"
```

Output:

```text
$ go run . "#0ea5e9"
--color-primary: oklch(68.469% 0.14787 237.323);
--color-primary-50: oklch(96.421% 0.03628 237.323);
--color-primary-100: oklch(92.798% 0.04912 237.323);
--color-primary-200: oklch(85.601% 0.07428 237.323);
--color-primary-300: oklch(79.038% 0.10088 237.323);
--color-primary-400: oklch(73.124% 0.12674 237.323);
--color-primary-500: oklch(68.469% 0.14787 237.323);
--color-primary-600: oklch(57.916% 0.12723 237.323);
--color-primary-700: oklch(46.152% 0.09843 237.323);
--color-primary-800: oklch(33.702% 0.06855 237.323);
--color-primary-900: oklch(20.024% 0.03274 237.323);
--color-primary-950: oklch(11.241% 0.01598 237.323);

```

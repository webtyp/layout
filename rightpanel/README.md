# RightPanel Layout

## Distribution

`RightPanel` implements a grid-based distribution inspired by the pa100t shell:

1. **Main Article** (left): the content, starting at the top of the frame. It has no
   title band: the module's name is the `platformd` header's `<h1>`.
2. **Aside Panel** (right): controls band (scope, filter, actions), content and footer.

## Styling

Uses semantic tokens from `webtyp.com/css` for all colors and spacing.
Proportions are controlled via `--rp-main-width` and `--rp-aside-width`.

---
PLAN: "fix: menu and drawer scroll correctly"
EXECUTOR: jules
REVIEWER: none
---

## Steps
1. In `platformd/css.go` for `Part(widget.Part("drawer-panel"))`, add `style.Scroll()`
2. In `platformd/css.go` for `Part(widget.Part("navbar"))`, add `style.Scroll()`
3. Ensure the test `TestMenu_Scrolls` passes.

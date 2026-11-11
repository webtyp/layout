---
PLAN: "fix: the side menu scrolls when it has more links than the window is tall"
EXECUTOR: jules
REVIEWER: none
STATUS: review
SESSION: 17455336479303599613
PR: https://github.com/webtyp/layout/pull/54
REVIEWED_COMMIT: 1238954015c58f876a55c858982490b0c961cfbc
---

## Steps
1. In `platformd/css.go` for `Part(widget.Part("drawer-panel"))`, add `style.Scroll()`
2. In `platformd/css.go` for `Part(widget.Part("navbar"))`, add `style.Scroll()`
3. Ensure the test `TestMenu_Scrolls` passes.

# Plan — Platformd Mobile App Bar

## Problem
In mobile view, `platformd` currently renders an empty fixed header while the hamburger button floats over the content inside `msg-stack` with a floating shadow. It looks disconnected (two separate pieces) and never auto-hides smoothly with the bar.

## Target
Integrate the hamburger button directly into `.pd__header`, add active module label/title to `.pd__header`, and apply the new `style.AppBar` recipe so the entire bar retracts cleanly on scroll down and re-appears on scroll up with scroll hysteresis.

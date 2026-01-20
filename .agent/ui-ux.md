---
description: UI/UX design guidelines
trigger: always_on
---

# UI/UX Guidelines

## Core Principles
- **Minimalist & Elegant:** Clean, enterprise-grade design focusing on content and readability.
- **Action-Oriented:** Prefer monochrome icons over text for actions.
- **Clean Code:** No inline styles, scripts, or HTML; use semantic tags and reusable components.
- **Responsive:** Desktop-first, 12-column grid system that is fully themeable.

## Visual Design
- **Layout:** Card-based design with generous whitespace and subtle shadows.
- **Aesthetics:** Flat UI with low contrast; no gradients, glossy effects, or clutter.
- **Color Palette:** Neutral gray background with a single muted accent color for primary actions.
- **Themes:** Native support for both Dark and Light modes.
- **Typography:** 
  - System font stack.
  - Semibold headings and section titles for clear hierarchy.
  - Regular body text.
- **Icons:** Font Awesome solid icons (`fas`) only; no custom SVGs or images.

## Components & Structure
- **Header (Fixed Top):** Contains title, navigation toggle, and pagination (prev/next).
- **Footer (Sticky Bottom):** Displays name, version, year, database name, and last build status.
- **Dashboard:** Organized list of panes/cards.
- **Page Data:** Structured layout containing header, footer, panes, and raw data.
- **Tab Navigation:** 
  - Hide default browser scrollbars for tab containers.
  - Use Font Awesome arrows (`fas fa-chevron-left`, `fas fa-chevron-right`) at the edges for horizontal navigation when content overflows.
  - Arrows should only appear when scrolling is available in that direction.

## Interactions & Technical
- **Dynamic Updates:** HTMX for AJAX-driven updates to tables and cards.
- **DOM Manipulation:** Lightweight jQuery for simple events and drag-and-drop.
- **Animations:** Minimal, subtle hover states only.
- **Validation:** Integrated formatters and validators for SQL and JSON inputs.
- **Frameworks:** Avoid heavy JS frameworks; stick to HTMX, jQuery, and CSS.

## Footer Implementation
- **Markup:** Use a semantic `<footer>` element with a fixed/sticky position at the bottom of the viewport.
- **Layout:** 
  - Use Flexbox (`display: flex`) for horizontal distribution.
  - Align items vertically in the center (`align-items: center`).
  - Use `justify-content: space-between` to separate the left (name/version), center (db name), and right (year/status) sections.
- **Styling:**
  - High `z-index` to ensure visibility over scrollable content.
  - Subtle top border (e.g., `1px solid`) to define the boundary.
  - Padding: Standardized small padding (e.g., `4px 12px`).
  - Font: Reduced font size (e.g., `11px` or `12px`) for a metadata-heavy look.
- **Responsiveness:** Maintain a single-line layout on desktop; consider stacking or hiding less critical info on mobile.

### Footer Data Context
- **Left Section:** Application Name, Version, and Author (e.g., `<i class="fas fa-robot"></i> Agent UI <i class="fas fa-code-branch"></i> v1.2.0 | <i class="fas fa-user"></i> Codeium`).
- **Center Section:** Active Database Name or Environment identifier (e.g., `<i class="fas fa-database"></i> DB: production_primary`).
- **Right Section:** Current Year and Last Build Status (e.g., `<i class="fas fa-calendar-alt"></i> 2024 | <i class="fas fa-circle"></i> Build: Success`).
- **Status Indicators:** Use a small dot icon (`fas fa-circle`) with semantic colors (success/danger) for build status.

# Frontend UX Optimization Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 36 UX/Accessibility issues across 5 priority levels in the web app

**Architecture:** Single worktree `feature/ux-optimization` isolated from main development. Changes organized by P0 (critical accessibility) → P1 (high UX) → P2 (medium) → P3 (low).

**Tech Stack:** Next.js, React, CSS Modules, Tailwind

---

## P0 Tasks (Critical - 8 items)

### Task P0-1: Add :focus-visible styles to globals.css
**Files:**
- Modify: `apps/web/app/globals.css`

**Step 1: Add :focus-visible CSS rules**
Add after existing :focus rules:
```css
:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

---

### Task P0-2: Add cursor-pointer to sidebar-link
**Files:**
- Modify: `apps/web/app/globals.css` (sidebar-link class)

**Step 1: Add cursor-pointer to .sidebar-link:hover**
```css
.sidebar-link:hover {
  background-color: var(--sidebar-link-hover-bg);
  cursor-pointer;
}
```

---

### Task P0-3: Add prefers-reduced-motion support
**Files:**
- Modify: `apps/web/app/globals.css`

**Step 1: Add reduced motion media query**
At end of CSS file:
```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}
```

---

### Task P0-4: Fix color contrast - muted text
**Files:**
- Modify: `apps/web/app/globals.css` (--muted variable)

**Step 1: Change --muted to accessible color**
```css
--muted: #475569;  /* was #64748b - 2.9:1 ratio, now ~7:1 */
```

---

### Task P0-5: Fix placeholder color contrast
**Files:**
- Modify: `apps/web/app/globals.css` (form-input::placeholder)

**Step 1: Update placeholder color**
```css
.form-input::placeholder {
  color: #94a3b8;  /* slate-400, accessible */
}
```

---

### Task P0-6: Search and add aria-label to icon buttons
**Files:**
- Grep: `apps/web/components/`, `apps/web/app/components/`
- Modify: All files with icon-only buttons

**Step 1: Search for icon buttons without aria-label**
Search pattern: `svg` (icon buttons) and check parent context

---

### Task P0-7: Add scope attribute to table headers
**Files:**
- Modify: `apps/web/app/components/SEHManager.tsx` (table headers)
- Modify: `apps/web/app/components/SEHRunsList.tsx` (table headers)

**Step 1: Add scope="col" to th elements**

---

### Task P0-8: Fix sidebar touch target sizes
**Files:**
- Modify: `apps/web/app/globals.css` (.sidebar-link)

**Step 1: Increase sidebar-link padding**
```css
.sidebar-link {
  padding: 12px 16px;  /* was ~8px - now ~48px height */
  min-height: 44px;
}
```

---

## P1 Tasks (High - 11 items)

### Task P1-1: Create Skeleton.tsx component
**Files:**
- Create: `apps/web/app/components/ui/Skeleton.tsx`

**Step 1: Create Skeleton component**
```tsx
export function Skeleton({ className = '' }: { className?: string }) {
  return (
    <div className={`animate-pulse bg-gray-200 rounded ${className}`} />
  );
}
```

---

### Task P1-2: Replace spinners in SEHRunsList with skeletons
**Files:**
- Modify: `apps/web/app/components/SEHRunsList.tsx`

**Step 1: Import Skeleton, replace spinner with skeleton rows**

---

### Task P1-3: Replace spinners in SEHDatasetsList with skeletons
**Files:**
- Modify: `apps/web/app/components/SEHDatasetsList.tsx`

---

### Task P1-4: Replace spinners in SkillsPageClient with skeletons
**Files:**
- Modify: `apps/web/app/components/SkillsPageClient.tsx`

---

### Task P1-5: Replace spinners in ExecutionsPageClient with skeletons
**Files:**
- Modify: `apps/web/app/components/ExecutionsPageClient.tsx`

---

### Task P1-6: Make AdminHeader sticky
**Files:**
- Modify: `apps/web/components/AdminHeader.tsx`

**Step 1: Add sticky positioning**
```tsx
<header className="sticky top-0 z-40 bg-white border-b">
```

---

### Task P1-7: Create Button.tsx component
**Files:**
- Create: `apps/web/app/components/ui/Button.tsx`

**Step 1: Create Button with variants and loading state**
```tsx
interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'destructive';
  isLoading?: boolean;
  children: React.ReactNode;
}
```

---

### Task P1-8: Add required field markers to all forms
**Files:**
- Modify: `CreateTaskForm.tsx`, `CreateExperimentForm.tsx`, `CreateDatasetForm.tsx`, `CreateMetricForm.tsx`, `CreateSkillForm.tsx`, `EditTaskForm.tsx`, `RunEvaluationForm.tsx`, `TriggerExecutionForm.tsx`, `CompareExperimentsForm.tsx`

**Step 1: Add asterisk and required indicator to label**
```tsx
<label htmlFor="name">
  Name <span className="text-red-500">*</span>
</label>
```

---

### Task P1-9: Add skip links
**Files:**
- Modify: `apps/web/app/layout.tsx`

**Step 1: Add skip link at top of layout**
```tsx
<a href="#main-content" className="skip-link">
  Skip to main content
</a>
```

---

### Task P1-10: Create 404 page
**Files:**
- Create: `apps/web/app/not-found.tsx`

**Step 1: Create user-friendly 404 page**

---

### Task P1-11: Create ErrorBoundary component
**Files:**
- Create: `apps/web/app/components/ErrorBoundary.tsx`
- Modify: `apps/web/app/layout.tsx`

**Step 1: Create ErrorBoundary with fallback UI**

---

## P2 Tasks (Medium - 13 items)

### Task P2-1: Unify transition durations
**Files:**
- Modify: `apps/web/app/globals.css`

**Step 1: Search and replace inconsistent transition durations**

---

### Task P2-2: Systematize z-index scale
**Files:**
- Modify: `apps/web/app/globals.css`

**Step 1: Define z-index scale**
```css
:root {
  --z-dropdown: 10;
  --z-sticky: 20;
  --z-modal: 30;
  --z-toast: 40;
}
```

---

### Task P2-3: Make Breadcrumb current page non-clickable
**Files:**
- Modify: `apps/web/components/Breadcrumb.tsx`

---

### Task P2-4: Add inline errors to CreateTaskForm
**Files:**
- Modify: `apps/web/app/components/CreateTaskForm.tsx`

---

### Task P2-5: Add inline errors to CreateMetricForm
**Files:**
- Modify: `apps/web/app/components/CreateMetricForm.tsx`

---

### Task P2-6: Add inline errors to EditTaskForm
**Files:**
- Modify: `apps/web/app/components/EditTaskForm.tsx`

---

### Task P2-7: Add success toast to all forms (after submit)
**Files:**
- Modify: All Create*/Edit* forms

---

### Task P2-8: Add empty state to SEHRunsList
**Files:**
- Modify: `apps/web/app/components/SEHRunsList.tsx`

---

### Task P2-9: Add empty state to SEHDatasetsList
**Files:**
- Modify: `apps/web/app/components/SEHDatasetsList.tsx`

---

### Task P2-10: Add clear filter button to filterable lists
**Files:**
- Modify: `SEHRunsList.tsx`, `SEHDatasetsList.tsx`, `SkillsPageClient.tsx`

---

### Task P2-11: Fix UserDropdown keyboard accessibility
**Files:**
- Modify: `apps/web/app/components/UserDropdown.tsx`

---

### Task P2-12: Add sort indicators to tables
**Files:**
- Modify: `SEHManager.tsx`, `SEHRunsList.tsx`, `SEHDatasetsList.tsx`

---

### Task P2-13: Add labels with for attribute to CreateTaskForm
**Files:**
- Modify: `apps/web/app/components/CreateTaskForm.tsx`

---

## P3 Tasks (Low - 4 items)

### Task P3-1: Add dark mode support
**Files:**
- Modify: `apps/web/app/globals.css`

---

### Task P3-2: Fix table row hover states
**Files:**
- Modify: Various list components

---

### Task P3-3: Fix hover scale causing layout shift
**Files:**
- Modify: `apps/web/app/globals.css`

---

### Task P3-4: Add BulkActionBar row selection to lists
**Files:**
- Modify: `SEHRunsList.tsx`, `SEHDatasetsList.tsx`, `SkillsPageClient.tsx`

---

## Execution Order

1. **P0 tasks first** - Critical accessibility fixes
2. **P1 tasks second** - High priority UX improvements
3. **P2 tasks third** - Medium priority polish
4. **P3 tasks last** - Low priority enhancements

**Parallelization opportunities:**
- P0-1 through P0-5: Can run in parallel (different CSS rules)
- P1-2 through P1-5: Can run in parallel (different list components)
- P2-4 through P2-6: Can run in parallel (different form files)
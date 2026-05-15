# shadcn/ui Component Rules — Incorrect / Correct

Pulled from official shadcn/ui skill. Load when writing UI components.

---

## Styling

### Spacing — use gap-*, not space-*

```tsx
// ❌ Incorrect
<div className="space-y-4">
  <Input />
  <Button />
</div>

// ✅ Correct
<div className="flex flex-col gap-4">
  <Input />
  <Button />
</div>
```

### Equal dimensions — use size-*, not w-* h-*

```tsx
// ❌ Incorrect
<Avatar className="w-10 h-10" />

// ✅ Correct
<Avatar className="size-10" />
```

### Colors — semantic tokens, never raw Tailwind colors

```tsx
// ❌ Incorrect
<span className="text-blue-600 bg-blue-50">Active</span>

// ✅ Correct
<Badge variant="secondary">Active</Badge>
// or
<span className="text-primary bg-primary/10">Active</span>
```

### No manual dark: overrides

```tsx
// ❌ Incorrect
<div className="bg-white dark:bg-gray-900">

// ✅ Correct
<div className="bg-background">
```

---

## Forms

### Form layout — FieldGroup + Field, not div + Label

```tsx
// ❌ Incorrect
<div className="space-y-4">
  <div>
    <label>Email</label>
    <Input />
  </div>
</div>

// ✅ Correct
<FieldGroup>
  <Field>
    <FieldLabel htmlFor="email">Email</FieldLabel>
    <Input id="email" />
  </Field>
</FieldGroup>
```

### Validation state

```tsx
// ❌ Incorrect
<Input className="border-red-500" />
<p className="text-red-500">Required</p>

// ✅ Correct
<Field data-invalid>
  <FieldLabel>Email</FieldLabel>
  <Input aria-invalid />
  <FieldDescription>Invalid email address.</FieldDescription>
</Field>
```

### Multiple choices — ToggleGroup, not looped Buttons

```tsx
// ❌ Incorrect
{options.map(opt => (
  <Button
    key={opt}
    variant={selected === opt ? "default" : "outline"}
    onClick={() => setSelected(opt)}
  >{opt}</Button>
))}

// ✅ Correct
<ToggleGroup type="single" value={selected} onValueChange={setSelected}>
  {options.map(opt => (
    <ToggleGroupItem key={opt} value={opt}>{opt}</ToggleGroupItem>
  ))}
</ToggleGroup>
```

---

## Component Composition

### Card structure — use all sub-components

```tsx
// ❌ Incorrect
<Card>
  <CardContent>
    <h3>Title</h3>
    <p>Everything dumped here</p>
    <Button>Action</Button>
  </CardContent>
</Card>

// ✅ Correct
<Card>
  <CardHeader>
    <CardTitle>Title</CardTitle>
    <CardDescription>Subtitle</CardDescription>
  </CardHeader>
  <CardContent>
    <p>Body content</p>
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>
```

### Icons in Button — use data-icon, no size classes

```tsx
// ❌ Incorrect
<Button>
  <SearchIcon className="w-4 h-4 mr-2" />
  Search
</Button>

// ✅ Correct
<Button>
  <SearchIcon data-icon="inline-start" />
  Search
</Button>
```

### Dialog — always needs a Title

```tsx
// ❌ Incorrect (accessibility violation)
<Dialog>
  <DialogContent>
    <p>Content without title</p>
  </DialogContent>
</Dialog>

// ✅ Correct
<Dialog>
  <DialogContent>
    <DialogTitle>Confirm Action</DialogTitle>
    <p>Are you sure?</p>
  </DialogContent>
</Dialog>
```

### Use semantic components, not custom markup

```tsx
// ❌ Incorrect
<div className="animate-pulse bg-gray-200 h-4 w-full rounded" />
<div className="border-t border-gray-200 my-4" />
<div className="inline-flex items-center rounded-full px-2 py-1 text-xs bg-green-100">Active</div>

// ✅ Correct
<Skeleton className="h-4 w-full" />
<Separator />
<Badge variant="secondary">Active</Badge>
```

---

## Runtime Context — Detect Project Config

Before adding components or running CLI commands:

```
!`bunx --bun shadcn@latest info --json`
```

This returns the installed components, aliases, framework, Tailwind version, icon library, and package manager. Use it to:
- Confirm correct import alias (`@/` vs `@rp/ui/`)
- Know which icon library to use (`lucide-react` vs `@tabler/icons-react`)
- Confirm Tailwind version (v4 uses `@theme`, v3 uses `tailwind.config.js`)
- Use correct package manager for any `bun add` commands

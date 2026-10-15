# Google-derived reference

Use this reference for technical artifacts or detailed style questions. Apply only the rules relevant to the requested destination.

## Core prose

- Write in US English unless the user or destination requires another variety.
- Use a conversational, friendly, respectful tone without slang, frivolity, or forced entertainment.
- Prefer active voice. Use passive voice when the actor is irrelevant or the object needs emphasis.
- Use present tense for general behavior. Use future tense only for a genuinely later action.
- Address the reader as **you**. Reserve **we** for a clearly named organization.
- Use familiar words in their primary sense. Avoid phrasal verbs, idioms, figurative language, and anthropomorphism when direct language is clearer.
- Include articles and helper words that improve comprehension. Do not omit them to sound terse.
- Use singular **they** when a gender-neutral pronoun is needed.
- Use common two-word contractions, especially negative contractions, when they make scanning easier.

## Claims, time, and sources

- Avoid superlatives and absolutes such as *best*, *simplest*, *fastest*, *never*, *always*, *ensure*, and *guarantee* unless the source proves them.
- Describe intended benefits or design instead of promising outcomes.
- Do not document unapproved future features or pre-announce changes.
- Avoid time anchors such as *currently*, *new*, *latest*, and *soon* in durable documentation unless a date or version makes them meaningful.
- Paraphrase and link to third-party material instead of copying it. Preserve required attribution and licenses.

## Structure

- Use sentence case for titles and headings.
- Use a task verb for a task heading and a noun phrase for a conceptual heading.
- Keep heading levels hierarchical and descriptive. Do not use headings only for visual styling.
- Keep one main idea per paragraph. Put important information first.
- Introduce a list or table with a complete sentence.
- Use numbered lists for sequences, bullets for unordered parallel items, and description lists for term-description pairs.
- Keep list items grammatically parallel. Do not create a list with one item.
- Use a table when readers must compare several properties across items. Do not use tables for layout or long prose.
- Use a notice only when information sits outside the main flow. Put prerequisites and required actions in the main text.

## Procedures and UI

- Write a procedure as numbered, imperative steps.
- Put context before the action: "To save the file, select **Save**."
- Prefer one action per step. Combine tightly coupled actions only when separating them would add friction.
- Mark a nonrequired step with **Optional:**.
- Include a result or reason only when it helps the reader verify or decide.
- Focus on the reader's goal. Mention a UI control when the path would otherwise be unclear.
- Bold visible UI labels and match their capitalization exactly.
- Avoid directional cues such as "on the right" when a label or accessible name identifies the control.

## Code, commands, and technical tokens

- Put code-related identifiers, filenames, commands, and literal input in code font.
- Keep product names, code, filenames, command syntax, and UI labels exact even when they break a general rule.
- Introduce a code block with a complete sentence. Use a colon when the sample follows immediately.
- Make command examples runnable when practical. Do not include a shell prompt unless it helps distinguish input from output.
- Use descriptive uppercase placeholders such as `PROJECT_ID`, and explain every placeholder.
- For API reference descriptions, state what the member does: "Creates a task," not "Create a task."
- Describe parameters by purpose and accepted values. Describe return values by the information they provide.

## Formatting and punctuation

- Use the serial comma.
- Prefer periods to semicolons. Use parentheses sparingly and keep them short.
- Use descriptive link text that tells readers what they will find. Avoid bare URLs and vague labels such as "click here."
- Use unambiguous dates: `January 19, 2026` in prose or `2026-01-19` when a numeric-only format is required.
- Spell out zero through nine in ordinary prose; use numerals for 10 and greater and for technical quantities, measurements, versions, percentages, and dimensions.
- Preserve the destination's established Markdown or HTML conventions. Use semantic HTML rather than visual tags when semantics matter.

## Accessibility, global readers, and examples

- Do not rely on color, position, sound, or another sensory cue alone.
- Give meaningful images concise alt text. Use empty alt text for decorative images.
- Do not use images of text, code, or terminal output when real text works.
- Avoid ableist, violent, unnecessarily gendered, and culturally specific language.
- Use examples that do not expose personally identifiable information. Prefer reserved domains, addresses, IP ranges, and phone numbers.
- Keep terminology, sentence patterns, capitalization, and formatting consistent to reduce ambiguity for readers and translators.

## Official source

This file is a high-frequency synthesis, not a complete reproduction of the [Google Developer Documentation Style Guide](https://developers.google.com/style). For a disputed word, specialized format, detailed compliance request, or exception, use [official-index.md](official-index.md) to find the relevant live page. Apply only the applicable category. If browsing isn't available, don't claim complete Google Style Guide compliance.

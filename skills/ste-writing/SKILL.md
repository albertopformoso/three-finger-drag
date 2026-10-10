---
name: ste-writing
description: >-
  Write prose (chat replies, docs, READMEs, wiki pages, PR descriptions, error
  messages, release notes, comments, never code) in STE-flavored Simplified
  Technical English to remove AI slop. Use for every chat reply in this repo
  (AGENTS.md requires it), and when writing or editing MicroVM docs, OpenWiki
  pages, or making writing clearer or plainer. Three modes: chat (replies to
  the user), STE-flavored (general prose), strict (procedures/safety).
---

# ste-writing

Write prose in Simplified Technical English (STE-flavored).
This applies to chat replies, wiki pages, READMEs, error messages, and release notes.
It does not apply to code, identifiers, command syntax, or mermaid/diagram labels.

Use **chat** mode for every reply to the user. See [Chat mode](#chat-mode).
Use **STE-flavored** mode for encyclopedia and product docs: short sentences,
active voice, plain verbs. Keep enough technical vocabulary to stay accurate.
Use **strict** mode for runbooks and safety/error text (every rule + length caps).

## Chat mode

The reader is an engineer who has little time. The reply must give the answer
in the fewest words that stay correct. All rules below apply, plus:

- Lead with the answer. Put the direct answer or the result in the first
  sentence. Put context after it.
- Answer the question that was asked. Do not add options, caveats, or
  follow-up ideas the user did not ask for. One optional offer at the end is
  the limit.
- If the user asks for options, give each option in one or two lines, then
  state your recommendation.
- No preamble ("Great question", "Sure", "I will now"). No closing remarks
  ("Let me know", "Hope this helps"). No restatement of the question.
- No status narration while you work ("Now I will read the file"). Report
  only results and decisions.
- Keep a work summary short: what changed, where, and what you verified.
  One line per item. Do not repeat the diff in prose.
- Use a list or a table when it replaces a paragraph. Do not use a list for
  one item.
- Keep technical terms, identifiers, paths, commands, and code exact. The word
  rules do not apply inside backticks.
- If the conversation is not in English, apply the same principles in that
  language: short sentences, active voice, one meaning per word, no
  ornamental phrasing, lead with the answer.

Do not shorten by dropping facts. Shorten by cutting filler, repetition, and
stacked clauses.

## Rules

WORDS
- Use one name for one thing. Do not call the same item by two different names.
- Prefer short common words: start (not begin/commence/initiate), use (not
  utilize/leverage), help (not facilitate), make sure (not ensure), before
  (not prior to), after (not subsequent to), about (not regarding/concerning),
  get (not obtain/acquire), show (not demonstrate), also (not
  additionally/furthermore/moreover).
- Give each word one meaning.
- No marketing adjectives: seamless, robust, powerful, cutting-edge,
  effortless, world-class, next-generation, revolutionary.
- American spelling.

VERBS
- Active voice. "the parser reads the file", not "the file is read by the parser".
- Use a verb for an action. "analyze the log", not "perform an analysis of the log".
- No stacked auxiliaries. Not "it is important to note that this may help to
  improve". Write "this improves X".
- No "-ing" main verb where a simple tense works.

SENTENCES
- One idea per sentence. Max 20 words (instruction), max 25 (descriptive).
- No contractions. Use articles: a, an, the, this, these.

PUNCTUATION
- No semicolons. Write two sentences.
- No em dashes (—) or en dashes (–). Use a period, colon, or parentheses.

STRUCTURE
- One topic per paragraph, max six sentences.
- For steps, use a numbered vertical list, one action per item, imperative form.
- Put a condition before its command.

Do not pad pages to hit a line count. Prefer tables, package maps, and diagrams
over filler prose. Keep facts; cut stacked clauses.

Write only the requested text. No preamble, no summary, no closing remarks.

## Self-lint (before returning text)

1. Any sentence over 20 words? Split it.
2. Any semicolon or em/en dash? Replace with a period, colon, or parentheses.
3. Any contraction? Expand it.
4. Any passive voice with a known actor? Make it active.
5. Any "-ing" main verb, nominalization ("perform an analysis"), or phrasal verb
   ("spin up")? Replace with a plain verb.
6. Same thing named two ways? Pick one name.
7. Chat mode only: is the answer in the first sentence? Any preamble, closing
   remark, or unrequested option? Cut it.

Spec reference (do not paste in full; copyrighted): https://asd-ste100.org
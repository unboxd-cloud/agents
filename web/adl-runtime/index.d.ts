// Type definitions for the ADL runtime.
//
// The runtime itself is the Go implementation in pkg/adl, compiled to
// WebAssembly (adl.wasm). These types describe the JSON it returns so the
// Langium tooling can consume the shared runtime with full type safety instead
// of maintaining a second parser.

export interface Position {
  line: number;
  column: number;
  offset: number;
}

export type Severity = "error" | "warning";

export interface Diagnostic {
  severity: Severity;
  message: string;
  pos: Position;
}

export interface Reference {
  name: string;
  /** Qualified name of the entity this reference bound to (empty if unresolved). */
  resolved?: string;
  pos: Position;
}

export interface TypeRef {
  primitive?: string;
  ref?: Reference;
  pos: Position;
}

export interface Field {
  name: string;
  type: TypeRef;
  required: boolean;
  pos: Position;
}

export interface Term {
  name: string;
  meaning?: string;
  mapping?: string;
  pos: Position;
}

export interface Rule {
  name: string;
  statement: string;
  pos: Position;
}

export interface Param {
  name: string;
  type: TypeRef;
  pos: Position;
}

interface DeclBase {
  pos: Position;
}

export interface Namespace extends DeclBase { kind: "Namespace"; name: string; }
export interface Vocabulary extends DeclBase { kind: "Vocabulary"; name: string; terms: Term[]; }
export interface Entity extends DeclBase { kind: "Entity"; name: string; super?: Reference; fields: Field[]; }
export interface Relation extends DeclBase { kind: "Relation"; name: string; source: Reference; target: Reference; fields: Field[]; }
export interface Brain extends DeclBase { kind: "Brain"; name: string; owns: string[]; }
export interface Mind extends DeclBase { kind: "Mind"; name: string; subject: Reference; fields: Field[]; }
export interface Belief extends DeclBase { kind: "Belief"; name: string; subject: Reference; claim: string; confidence?: number; source?: string; }
export interface Constitution extends DeclBase { kind: "Constitution"; name: string; rules: Rule[]; }
export interface Policy extends DeclBase { kind: "Policy"; name: string; rules: Rule[]; }
export interface Api extends DeclBase { kind: "Api"; method: string; path: string; target: string; }
export interface Function extends DeclBase { kind: "Function"; name: string; params: Param[]; return: TypeRef; target: string; }
export interface SurrealMlBinding extends DeclBase { kind: "SurrealMlBinding"; name: string; purpose: string; input: Reference; output: Reference; }

export type Declaration =
  | Namespace
  | Vocabulary
  | Entity
  | Relation
  | Brain
  | Mind
  | Belief
  | Constitution
  | Policy
  | Api
  | Function
  | SurrealMlBinding;

export interface Model {
  declarations: Declaration[];
}

export interface Result {
  model: Model | null;
  diagnostics: Diagnostic[];
}

export interface ADLRuntime {
  /** Lex, parse, and validate (resolve references). The canonical pipeline. */
  compile(source: string): Result;
  /** Lex and parse only (syntax), without reference resolution. */
  parse(source: string): Result;
}

/**
 * Load the ADL runtime. Pass the URL/path to adl.wasm; defaults to the
 * adl.wasm sitting next to this module. Works in Node and the browser.
 */
export function loadADL(wasmUrl?: string | URL): Promise<ADLRuntime>;

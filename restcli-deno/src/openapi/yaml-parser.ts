/**
 * Minimal YAML Parser for OpenAPI Specs
 * Built from scratch - no external dependencies
 *
 * Supports:
 * - Objects (key: value)
 * - Arrays (- item)
 * - Strings (quoted and unquoted)
 * - Numbers, booleans, null
 * - Multi-line strings (| and >)
 * - Indentation-based structure
 *
 * Does NOT support:
 * - Anchors and aliases (&, *)
 * - Multiple documents (---)
 * - Complex types
 * - Custom tags
 */

type YamlValue = string | number | boolean | null | YamlObject | YamlArray;
type YamlObject = { [key: string]: YamlValue };
type YamlArray = YamlValue[];

interface Line {
  indent: number;
  content: string;
  lineNumber: number;
}

export function parseYaml(input: string): YamlValue {
  const lines = preprocessLines(input);

  if (lines.length === 0) {
    return null;
  }

  const { value } = parseValue(lines, 0, -1);
  return value;
}

/**
 * Preprocess input into lines with indentation info
 */
function preprocessLines(input: string): Line[] {
  const rawLines = input.split('\n');
  const lines: Line[] = [];

  for (let i = 0; i < rawLines.length; i++) {
    const line = rawLines[i];

    // Skip empty lines and comments
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith('#')) {
      continue;
    }

    // Skip document markers
    if (trimmed === '---' || trimmed === '...') {
      continue;
    }

    // Calculate indentation (number of leading spaces)
    const indent = line.length - line.trimStart().length;

    lines.push({
      indent,
      content: trimmed,
      lineNumber: i + 1,
    });
  }

  return lines;
}

/**
 * Parse a value starting at the given index
 * Returns the parsed value and the next index to process
 */
function parseValue(
  lines: Line[],
  index: number,
  parentIndent: number
): { value: YamlValue; nextIndex: number } {
  if (index >= lines.length) {
    return { value: null, nextIndex: index };
  }

  const line = lines[index];

  // Check if this is an array item
  if (line.content.startsWith('- ')) {
    return parseArray(lines, index, parentIndent);
  }

  // Check if this is an object (has a key)
  if (line.content.includes(':')) {
    return parseObject(lines, index, parentIndent);
  }

  // Scalar value
  return { value: parseScalar(line.content), nextIndex: index + 1 };
}

/**
 * Parse an object
 */
function parseObject(
  lines: Line[],
  index: number,
  parentIndent: number
): { value: YamlObject; nextIndex: number } {
  const obj: YamlObject = {};
  let currentIndex = index;
  const currentIndent = lines[currentIndex].indent;

  while (currentIndex < lines.length) {
    const line = lines[currentIndex];

    // Stop if we've dedented
    if (line.indent < currentIndent) {
      break;
    }

    // Skip if more indented (child of previous key)
    if (line.indent > currentIndent) {
      currentIndex++;
      continue;
    }

    // Parse key-value pair
    const colonIndex = line.content.indexOf(':');
    if (colonIndex === -1) {
      currentIndex++;
      continue;
    }

    const key = line.content.substring(0, colonIndex).trim();
    const valueStr = line.content.substring(colonIndex + 1).trim();

    if (valueStr) {
      // Inline value
      obj[key] = parseScalar(valueStr);
      currentIndex++;
    } else {
      // Value on next line(s)
      currentIndex++;

      if (currentIndex >= lines.length) {
        obj[key] = null;
        break;
      }

      const nextLine = lines[currentIndex];

      // Check for multi-line string
      if (valueStr === '|' || valueStr === '>') {
        const { value, nextIndex } = parseMultilineString(
          lines,
          currentIndex,
          line.indent,
          valueStr === '>'
        );
        obj[key] = value;
        currentIndex = nextIndex;
      } else if (nextLine.indent > line.indent) {
        // Nested value
        const { value, nextIndex } = parseValue(lines, currentIndex, line.indent);
        obj[key] = value;
        currentIndex = nextIndex;
      } else {
        obj[key] = null;
      }
    }
  }

  return { value: obj, nextIndex: currentIndex };
}

/**
 * Parse an array
 */
function parseArray(
  lines: Line[],
  index: number,
  parentIndent: number
): { value: YamlArray; nextIndex: number } {
  const arr: YamlArray = [];
  let currentIndex = index;
  const currentIndent = lines[currentIndex].indent;

  while (currentIndex < lines.length) {
    const line = lines[currentIndex];

    // Stop if we've dedented
    if (line.indent < currentIndent) {
      break;
    }

    // Skip if more indented (child of previous item)
    if (line.indent > currentIndent) {
      currentIndex++;
      continue;
    }

    // Must be array item
    if (!line.content.startsWith('- ')) {
      break;
    }

    const valueStr = line.content.substring(2).trim();

    if (valueStr) {
      // Inline value
      arr.push(parseScalar(valueStr));
      currentIndex++;
    } else {
      // Value on next line(s)
      currentIndex++;

      if (currentIndex >= lines.length) {
        arr.push(null);
        break;
      }

      const nextLine = lines[currentIndex];

      if (nextLine.indent > line.indent) {
        // Nested value
        const { value, nextIndex } = parseValue(lines, currentIndex, line.indent);
        arr.push(value);
        currentIndex = nextIndex;
      } else {
        arr.push(null);
      }
    }
  }

  return { value: arr, nextIndex: currentIndex };
}

/**
 * Parse a multi-line string (| or >)
 */
function parseMultilineString(
  lines: Line[],
  index: number,
  parentIndent: number,
  folded: boolean // true for >, false for |
): { value: string; nextIndex: number } {
  const stringLines: string[] = [];
  let currentIndex = index;

  while (currentIndex < lines.length) {
    const line = lines[currentIndex];

    // Stop if we've dedented or back to parent level
    if (line.indent <= parentIndent) {
      break;
    }

    // Add the content (remove the base indentation)
    const baseIndent = lines[index]?.indent || 0;
    const relativeIndent = line.indent - baseIndent;
    const padding = ' '.repeat(relativeIndent);
    stringLines.push(padding + line.content);
    currentIndex++;
  }

  if (folded) {
    // > (folded): join lines with spaces, preserve double newlines as paragraph breaks
    return { value: stringLines.join(' '), nextIndex: currentIndex };
  } else {
    // | (literal): preserve newlines
    return { value: stringLines.join('\n'), nextIndex: currentIndex };
  }
}

/**
 * Parse a scalar value (string, number, boolean, null)
 */
function parseScalar(value: string): string | number | boolean | null {
  // Null
  if (value === 'null' || value === '~' || value === '') {
    return null;
  }

  // Boolean
  if (value === 'true' || value === 'yes' || value === 'on') {
    return true;
  }
  if (value === 'false' || value === 'no' || value === 'off') {
    return false;
  }

  // Number
  const num = Number(value);
  if (!isNaN(num) && value.trim() === String(num)) {
    return num;
  }

  // String - remove quotes if present
  if ((value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))) {
    return value.slice(1, -1);
  }

  return value;
}

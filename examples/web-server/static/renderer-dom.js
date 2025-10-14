/**
 * DOMRenderer - DOM-based matrix renderer
 */
class DOMRenderer {
    constructor(container) {
        this.container = container;
        this.width = 0;
        this.height = 0;
        this.cells = [];
        this.rowElements = [];
    }

    initialize(width, height) {
        this.width = width;
        this.height = height;
        this.cells = Array(height).fill(null).map(() => Array(width).fill(null));
        this.rowElements = [];

        // Clear container
        this.container.innerHTML = '';
        this.container.classList.remove('loading');

        // Create rows
        for (let y = 0; y < height; y++) {
            const row = document.createElement('div');
            row.className = 'matrix-row';
            row.dataset.row = y;

            // Create cells for this row
            for (let x = 0; x < width; x++) {
                const span = document.createElement('span');
                span.className = 'matrix-cell';
                span.dataset.x = x;
                span.dataset.y = y;
                span.textContent = ' ';
                row.appendChild(span);
            }

            this.rowElements.push(row);
            this.container.appendChild(row);
        }
    }

    renderSnapshot(cellsData) {
        if (!cellsData || cellsData.length === 0) return;

        const height = cellsData.length;
        const width = cellsData[0]?.length || 0;

        // Initialize if dimensions changed
        if (this.width !== width || this.height !== height) {
            this.initialize(width, height);
        }

        // Render all cells
        for (let y = 0; y < height; y++) {
            for (let x = 0; x < width; x++) {
                const cell = cellsData[y][x];
                this.updateCell(x, y, cell);
            }
        }
    }

    renderDelta(deltaUpdates) {
        if (!deltaUpdates || deltaUpdates.length === 0) return;

        // Apply each cell update
        for (const update of deltaUpdates) {
            this.updateCell(update.x, update.y, update.cell);
        }
    }

    updateCell(x, y, cell) {
        if (x < 0 || x >= this.width || y < 0 || y >= this.height) {
            return;
        }

        const row = this.rowElements[y];
        if (!row) return;

        const span = row.children[x];
        if (!span) return;

        // Check if cell actually changed to avoid unnecessary DOM updates
        const oldCell = this.cells[y][x];
        if (oldCell &&
            oldCell.char === (cell.char || ' ') &&
            oldCell.fg === cell.fg &&
            oldCell.bg === cell.bg &&
            oldCell.style === cell.style) {
            return; // No change, skip DOM update
        }

        // Update character
        const char = cell.char || ' ';
        if (span.textContent !== char) {
            span.textContent = char;
        }

        // Build inline style
        let style = '';

        // Colors
        if (cell.fg) {
            style += `color: ${cell.fg}; `;
        } else {
            style += 'color: #fff; ';
        }

        if (cell.bg) {
            style += `background: ${cell.bg}; `;
        } else {
            style += 'background: #000; ';
        }

        // Styles (bitfield)
        const styleFlags = cell.style || 0;
        if (styleFlags & 1) { // Bold
            style += 'font-weight: bold; ';
        }
        if (styleFlags & 2) { // Italic
            style += 'font-style: italic; ';
        }
        if (styleFlags & 4) { // Underline
            style += 'text-decoration: underline; ';
        }
        // Note: Blink (8) and Reverse (16) are harder to implement in CSS

        span.style.cssText = style;

        // Store cell state for future comparison
        this.cells[y][x] = {
            char: char,
            fg: cell.fg,
            bg: cell.bg,
            style: cell.style
        };
    }

    clear() {
        const emptyCell = { char: ' ', fg: '', bg: '', style: 0 };
        for (let y = 0; y < this.height; y++) {
            for (let x = 0; x < this.width; x++) {
                this.updateCell(x, y, emptyCell);
            }
        }
    }
}

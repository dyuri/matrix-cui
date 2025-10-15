/**
 * CanvasRenderer - Canvas-based matrix renderer
 *
 * Performance optimized renderer using HTML5 Canvas 2D API.
 * Should be faster than DOM renderer for frequent updates.
 */
class CanvasRenderer {
    constructor(container) {
        this.container = container;
        this.width = 0;
        this.height = 0;
        this.cells = [];
        this.canvas = null;
        this.ctx = null;
        this.charWidth = 0;
        this.charHeight = 0;
        this.needsRedraw = false;
        this.dirtyRegion = null;
    }

    initialize(width, height) {
        this.width = width;
        this.height = height;
        this.cells = Array(height).fill(null).map(() => Array(width).fill(null));

        // Clear container and create canvas
        this.container.innerHTML = '';
        this.container.classList.remove('loading');

        this.canvas = document.createElement('canvas');
        this.canvas.className = 'matrix-canvas';
        this.canvas.style.display = 'block';
        this.canvas.style.imageRendering = 'crisp-edges'; // Sharper text
        this.ctx = this.canvas.getContext('2d', { alpha: false }); // No alpha for better performance

        // Measure character dimensions
        this.measureCharacter();

        // Set canvas size based on character dimensions
        this.canvas.width = this.charWidth * width;
        this.canvas.height = this.charHeight * height;

        // Set canvas display size (can be different from buffer size for scaling)
        this.canvas.style.width = `${this.canvas.width}px`;
        this.canvas.style.height = `${this.canvas.height}px`;

        this.container.appendChild(this.canvas);

        // Initial clear
        this.ctx.fillStyle = '#000';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
    }

    measureCharacter() {
        // Use same font as CSS
        const style = window.getComputedStyle(this.container);
        const fontSize = parseFloat(style.fontSize) || 16;
        const fontFamily = style.fontFamily || "'Courier New', Monaco, 'Lucida Console', monospace";

        this.ctx.font = `${fontSize}px ${fontFamily}`;
        this.ctx.textBaseline = 'top';

        // Measure a monospace character
        const metrics = this.ctx.measureText('M');
        this.charWidth = Math.ceil(metrics.width);
        this.charHeight = Math.ceil(fontSize * 1.2); // Match line-height from CSS
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

        // Check if cell actually changed
        const oldCell = this.cells[y][x];
        const char = cell.char || ' ';
        if (oldCell &&
            oldCell.char === char &&
            oldCell.fg === cell.fg &&
            oldCell.bg === cell.bg &&
            oldCell.style === cell.style) {
            return; // No change, skip rendering
        }

        // Store cell state
        this.cells[y][x] = {
            char: char,
            fg: cell.fg,
            bg: cell.bg,
            style: cell.style
        };

        // Draw the cell
        this.drawCell(x, y, cell);
    }

    drawCell(x, y, cell) {
        const px = x * this.charWidth;
        const py = y * this.charHeight;
        const char = cell.char || ' ';
        const fg = cell.fg || '#fff';
        const bg = cell.bg || '#000';
        const styleFlags = cell.style || 0;

        // Draw background
        this.ctx.fillStyle = bg;
        this.ctx.fillRect(px, py, this.charWidth, this.charHeight);

        // Skip drawing space characters (background is enough)
        if (char === ' ') return;

        // Set up text style
        const style = window.getComputedStyle(this.container);
        const fontSize = parseFloat(style.fontSize) || 16;
        const fontFamily = style.fontFamily || "'Courier New', Monaco, 'Lucida Console', monospace";

        let font = '';

        // Italic
        if (styleFlags & 2) {
            font += 'italic ';
        }

        // Bold
        if (styleFlags & 1) {
            font += 'bold ';
        }

        font += `${fontSize}px ${fontFamily}`;
        this.ctx.font = font;
        this.ctx.textBaseline = 'top';

        // Draw character
        this.ctx.fillStyle = fg;
        this.ctx.fillText(char, px, py);

        // Underline (manual drawing since canvas doesn't support it)
        if (styleFlags & 4) {
            const underlineY = py + this.charHeight - 2;
            this.ctx.strokeStyle = fg;
            this.ctx.lineWidth = 1;
            this.ctx.beginPath();
            this.ctx.moveTo(px, underlineY);
            this.ctx.lineTo(px + this.charWidth, underlineY);
            this.ctx.stroke();
        }
    }

    clear() {
        // Clear canvas
        this.ctx.fillStyle = '#000';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);

        // Clear cell state
        const emptyCell = { char: ' ', fg: '', bg: '', style: 0 };
        for (let y = 0; y < this.height; y++) {
            for (let x = 0; x < this.width; x++) {
                this.cells[y][x] = { ...emptyCell };
            }
        }
    }

    // Get the canvas element for mouse event handling
    getElement() {
        return this.canvas;
    }
}

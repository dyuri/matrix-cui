/**
 * Main application logic
 */

// Determine WebSocket URL
const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = `${wsProtocol}//${window.location.host}/ws`;

// Initialize protocol and renderer
const protocol = new MatrixProtocol(wsUrl);
const terminal = document.getElementById('terminal');
const renderer = new DOMRenderer(terminal);

// UI elements
const statusEl = document.getElementById('status');
const sizeEl = document.getElementById('size');

// Status updates
protocol.on('connected', () => {
    console.log('Connected to server');
    statusEl.textContent = 'Connected';
    statusEl.className = 'status-indicator connected';
});

protocol.on('disconnected', () => {
    console.log('Disconnected from server');
    statusEl.textContent = 'Disconnected';
    statusEl.className = 'status-indicator';
});

protocol.on('error', (data) => {
    console.error('Protocol error:', data.error);
    statusEl.textContent = 'Error';
    statusEl.className = 'status-indicator';
});

// Handle initial snapshot
protocol.on('snapshot', (data) => {
    console.log('Received snapshot:', data.width, 'x', data.height);
    renderer.initialize(data.width, data.height);
    renderer.renderSnapshot(data.cells);
    sizeEl.textContent = `${data.width}x${data.height}`;
});

// Handle resize events
protocol.on('resize', (data) => {
    console.log('Terminal resized:', data.width, 'x', data.height);
    renderer.initialize(data.width, data.height);
    sizeEl.textContent = `${data.width}x${data.height}`;
});

// Keyboard event handling
terminal.addEventListener('keydown', (e) => {
    e.preventDefault(); // Prevent browser shortcuts

    // Map key codes
    const key = mapKeyCode(e.key);
    const rune = e.key.length === 1 ? e.key.charCodeAt(0) : 0;

    // Send keyboard event to server
    protocol.sendEvent({
        event: 'key',
        key: key || 'None',
        rune: rune,
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });

    console.log('Key pressed:', e.key, key, rune);
});

// Mouse event handling
let isMouseDown = false;

terminal.addEventListener('mousedown', (e) => {
    e.preventDefault();
    terminal.focus();
    isMouseDown = true;

    const { x, y } = getMousePosition(e);

    // Send mouse press event to server
    protocol.sendEvent({
        event: 'mouse',
        x: x,
        y: y,
        button: getMouseButton(e.button),
        action: 'Press',
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });

    console.log('Mouse down at:', x, y);
});

terminal.addEventListener('mouseup', (e) => {
    isMouseDown = false;
    const { x, y } = getMousePosition(e);

    // Send mouse release event to server
    protocol.sendEvent({
        event: 'mouse',
        x: x,
        y: y,
        button: getMouseButton(e.button),
        action: 'Release',
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });

    console.log('Mouse up at:', x, y);
});

terminal.addEventListener('mousemove', (e) => {
    if (!isMouseDown) return;

    const { x, y } = getMousePosition(e);

    // Send mouse move event to server
    protocol.sendEvent({
        event: 'mouse',
        x: x,
        y: y,
        button: 'Left',  // Assume left button during drag
        action: 'Move',
        alt: e.altKey,
        ctrl: e.ctrlKey,
        shift: e.shiftKey
    });

    console.log('Mouse move to:', x, y);
});

// Helper function to get mouse position in matrix coordinates
function getMousePosition(e) {
    const rect = terminal.getBoundingClientRect();
    const style = window.getComputedStyle(terminal);
    const fontSize = parseFloat(style.fontSize);
    const charWidth = fontSize * 0.6; // Approximate character width
    const charHeight = fontSize * 1.2; // Line height

    const x = Math.floor((e.clientX - rect.left) / charWidth);
    const y = Math.floor((e.clientY - rect.top) / charHeight);

    return { x, y };
}

// Helper function to map key codes
function mapKeyCode(key) {
    const keyMap = {
        'Enter': 'Enter',
        'Escape': 'Escape',
        'Backspace': 'Backspace',
        'Tab': 'Tab',
        ' ': 'Space',
        'ArrowUp': 'Up',
        'ArrowDown': 'Down',
        'ArrowLeft': 'Left',
        'ArrowRight': 'Right',
        'Home': 'Home',
        'End': 'End',
        'PageUp': 'PageUp',
        'PageDown': 'PageDown',
        'Insert': 'Insert',
        'Delete': 'Delete',
        'F1': 'F1',
        'F2': 'F2',
        'F3': 'F3',
        'F4': 'F4',
        'F5': 'F5',
        'F6': 'F6',
        'F7': 'F7',
        'F8': 'F8',
        'F9': 'F9',
        'F10': 'F10',
        'F11': 'F11',
        'F12': 'F12',
    };
    return keyMap[key] || '';
}

// Helper function to map mouse button codes
function getMouseButton(button) {
    const buttonMap = {
        0: 'Left',
        1: 'Middle',
        2: 'Right'
    };
    return buttonMap[button] || 'Left';
}

// Focus terminal on load
terminal.focus();

// Mark terminal as loading
terminal.classList.add('loading');

// Connect to server
protocol.connect().catch((error) => {
    console.error('Failed to connect:', error);
    terminal.classList.remove('loading');
    terminal.innerHTML = `<div class="error-message">Failed to connect to server: ${error.message}</div>`;
});

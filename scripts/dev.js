#!/usr/bin/env node
/**
 * Cross-platform development server script
 * Provides clear feedback on process status
 */

const { spawn, execSync } = require('child_process');
const path = require('path');

// ANSI color codes
const colors = {
    reset: '\x1b[0m',
    bright: '\x1b[1m',
    dim: '\x1b[2m',
    green: '\x1b[32m',
    blue: '\x1b[34m',
    yellow: '\x1b[33m',
    cyan: '\x1b[36m',
    red: '\x1b[31m',
    magenta: '\x1b[35m',
};

function log(prefix, color, message) {
    const time = new Date().toLocaleTimeString('en-US', { hour12: false });
    console.log(`${colors.dim}[${time}]${colors.reset} ${color}[${prefix}]${colors.reset} ${message}`);
}

function logSystem(message) {
    console.log(`\n${colors.cyan}${colors.bright}► ${message}${colors.reset}\n`);
}

function logSuccess(message) {
    console.log(`${colors.green}${colors.bright}✓ ${message}${colors.reset}`);
}

function logError(message) {
    console.log(`${colors.red}${colors.bright}✗ ${message}${colors.reset}`);
}

function runSync(command, description) {
    logSystem(description);
    try {
        execSync(command, {
            stdio: 'inherit',
            shell: true,
            cwd: process.cwd()
        });
        logSuccess(`${description} - Complete`);
        return true;
    } catch (error) {
        logError(`${description} - Failed`);
        return false;
    }
}

function spawnProcess(name, command, args, color) {
    const isWindows = process.platform === 'win32';

    let proc;
    if (isWindows) {
        proc = spawn('cmd', ['/c', command, ...args], {
            cwd: process.cwd(),
            shell: true,
        });
    } else {
        proc = spawn(command, args, {
            cwd: process.cwd(),
            shell: true,
        });
    }

    proc.stdout.on('data', (data) => {
        const lines = data.toString().trim().split('\n');
        lines.forEach(line => {
            if (line.trim()) {
                log(name, color, line);
            }
        });
    });

    proc.stderr.on('data', (data) => {
        const lines = data.toString().trim().split('\n');
        lines.forEach(line => {
            if (line.trim()) {
                log(name, color, `${colors.yellow}${line}${colors.reset}`);
            }
        });
    });

    proc.on('error', (error) => {
        log(name, colors.red, `Error: ${error.message}`);
    });

    proc.on('exit', (code) => {
        if (code !== 0 && code !== null) {
            log(name, colors.red, `Process exited with code ${code}`);
        }
    });

    return proc;
}

async function main() {
    console.log('\n');
    console.log(`${colors.magenta}${colors.bright}╔════════════════════════════════════════════╗${colors.reset}`);
    console.log(`${colors.magenta}${colors.bright}║     Full-Stack Go Development Server       ║${colors.reset}`);
    console.log(`${colors.magenta}${colors.bright}╚════════════════════════════════════════════╝${colors.reset}`);
    console.log('\n');

    // Step 1: Install npm dependencies
    if (!runSync('npm install', 'Installing npm dependencies')) {
        process.exit(1);
    }

    // Step 2: Download Go modules
    if (!runSync('go mod download', 'Downloading Go modules')) {
        process.exit(1);
    }

    // Step 3: Copy vendor assets
    logSystem('Copying vendor assets...');
    const vendorDir = path.join('web', 'assets', 'vendor');

    // Create vendor directory if it doesn't exist
    if (require('fs').existsSync(vendorDir) === false) {
        require('fs').mkdirSync(vendorDir, { recursive: true });
    }

    // Copy HTMX
    if (!runSync(process.platform === 'win32' ?
        'copy node_modules\\htmx.org\\dist\\htmx.min.js web\\assets\\vendor\\' :
        'cp node_modules/htmx.org/dist/htmx.min.js web/assets/vendor/',
        'Copying HTMX')) {
        process.exit(1);
    }

    // Copy Alpine.js
    if (!runSync(process.platform === 'win32' ?
        'copy node_modules\\alpinejs\\dist\\cdn.min.js web\\assets\\vendor\\alpine.min.js' :
        'cp node_modules/alpinejs/dist/cdn.min.js web/assets/vendor/alpine.min.js',
        'Copying Alpine.js')) {
        process.exit(1);
    }

    // Copy Lucide
    if (!runSync(process.platform === 'win32' ?
        'copy node_modules\\lucide\\dist\\umd\\lucide.min.js web\\assets\\vendor\\' :
        'cp node_modules/lucide/dist/umd/lucide.min.js web/assets/vendor/',
        'Copying Lucide values')) {
        process.exit(1);
    }

    // Copy ECharts
    if (!runSync(process.platform === 'win32' ?
        'copy node_modules\\echarts\\dist\\echarts.min.js web\\assets\\vendor\\' :
        'cp node_modules/echarts/dist/echarts.min.js web/assets/vendor/',
        'Copying ECharts')) {
        process.exit(1);
    }

    logSystem('Starting development watchers...');
    console.log(`${colors.dim}Press Ctrl+C to stop all processes${colors.reset}\n`);

    // Start Go server with Air (hot reload)
    const goProc = spawnProcess(
        'GO ',
        'go',
        ['run', 'github.com/air-verse/air'],
        colors.green
    );

    // Start Tailwind CSS watcher
    const cssProc = spawnProcess(
        'CSS',
        'npx',
        ['@tailwindcss/cli', '-i', './web/assets/css/app.css', '-o', './web/assets/css/output.css', '--watch'],
        colors.blue
    );

    // Start Esbuild for React (Watch mode)
    const esbuildProc = spawnProcess(
        'JS ',
        'npx',
        [
            'esbuild',
            'web/assets/js/react/main.jsx',
            '--bundle',
            '--outfile=web/assets/js/react.bundle.js',
            '--loader:.jsx=jsx',
            '--sourcemap',
            '--watch'
        ],
        colors.yellow
    );

    // Build Highlight.js (Single run, not watched as it rarely changes)
    logSystem('Bundling Highlight.js...');
    try {
        execSync('npx esbuild web/assets/js/highlight_entry.js --bundle --minify --outfile=web/assets/vendor/highlight.js --format=iife', { stdio: 'inherit' });
        logSuccess('Bundled Highlight.js');
    } catch (e) {
        logError('Failed to bundle Highlight.js');
    }

    // Handle graceful shutdown
    const cleanup = () => {
        console.log(`\n${colors.yellow}Shutting down...${colors.reset}`);
        goProc.kill();
        cssProc.kill();
        esbuildProc.kill();
        process.exit(0);
    };

    process.on('SIGINT', cleanup);
    process.on('SIGTERM', cleanup);

    // Keep the script running
    goProc.on('exit', () => {
        cssProc.kill();
        esbuildProc.kill();
        process.exit(1);
    });
}

main().catch((error) => {
    logError(`Fatal error: ${error.message}`);
    process.exit(1);
});

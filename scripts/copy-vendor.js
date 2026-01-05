// Script to copy vendor files from node_modules to web/static/vendor
const fs = require('fs');
const path = require('path');

const vendorDir = path.join(__dirname, '..', 'web', 'static', 'vendor');

// Create vendor directory if it doesn't exist
if (!fs.existsSync(vendorDir)) {
    fs.mkdirSync(vendorDir, { recursive: true });
}

// Files to copy
const filesToCopy = [
    {
        src: 'node_modules/htmx.org/dist/htmx.min.js',
        dest: 'htmx.min.js'
    },
    {
        src: 'node_modules/alpinejs/dist/cdn.min.js',
        dest: 'alpine.min.js'
    },
    {
        src: 'node_modules/lucide/dist/umd/lucide.min.js',
        dest: 'lucide.min.js'
    }
];

console.log('Copying vendor files...');

filesToCopy.forEach(file => {
    const srcPath = path.join(__dirname, '..', file.src);
    const destPath = path.join(vendorDir, file.dest);

    try {
        fs.copyFileSync(srcPath, destPath);
        console.log(`  ✓ ${file.dest}`);
    } catch (err) {
        console.error(`  ✗ Failed to copy ${file.dest}: ${err.message}`);
    }
});

console.log('Vendor files copied to web/static/vendor/');

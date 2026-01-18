const fs = require('fs');
const path = require('path');

const copyFile = (src, dest) => {
    const srcPath = path.join(__dirname, '..', src);
    const destPath = path.join(__dirname, '..', dest);

    // Ensure destination directory exists
    const destDir = path.dirname(destPath);
    if (!fs.existsSync(destDir)) {
        fs.mkdirSync(destDir, { recursive: true });
    }

    fs.copyFileSync(srcPath, destPath);
    console.log(`Copied ${src} to ${dest}`);
};

const files = [
    { src: 'node_modules/htmx.org/dist/htmx.min.js', dest: 'web/assets/vendor/htmx.min.js' },
    { src: 'node_modules/alpinejs/dist/cdn.min.js', dest: 'web/assets/vendor/alpine.min.js' },
    { src: 'node_modules/lucide/dist/umd/lucide.min.js', dest: 'web/assets/vendor/lucide.min.js' },
    { src: 'node_modules/echarts/dist/echarts.min.js', dest: 'web/assets/vendor/echarts.min.js' }
];

try {
    files.forEach(file => copyFile(file.src, file.dest));
    console.log('Vendor assets copied successfully.');
} catch (err) {
    console.error('Error copying vendor assets:', err);
    process.exit(1);
}

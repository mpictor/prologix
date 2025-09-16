// web server frontend code, plotting 2d graph
const canvas = document.getElementById('chart');
const ctx = canvas.getContext('2d');
const refreshButton = document.getElementById('refresh');
let lastData = null;

// Function to fetch data from the backend
async function fetchData() {
    try {
        const response = await fetch('/data'); // Backend endpoint
        if (!response.ok) {
            throw new Error('Failed to fetch data');
        }
        return await response.json();
    } catch (error) {
        console.error('Error fetching data:', error);
        return null;
    }
}

function cssColor(context, id) {
    s = getComputedStyle(context.canvas)
    c = s.getPropertyValue("--" + id + "-color")
    return c
}
function cssWidth(context, id) {
    s = getComputedStyle(context.canvas)
    c = s.getPropertyValue("--" + id + "-width")
    return c
}

// Function to draw the 2D line chart with legends and grid
function drawChart(data) {
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    if (!data || !data.coordinates || data.coordinates.length === 0) {
        ctx.fillText('No data available', canvas.width / 2, canvas.height / 2);
        return;
    }

    const { coordinates, hLegend, hMult, vLegend, vMult } = data;

    // Define margins to shrink the graph
    const wmargin = 70; // Increased margin to accommodate rotated legend
    const hmargin = 50;
    const graphWidth = canvas.width - wmargin * 2;
    const graphHeight = canvas.height - hmargin * 2;

    // Find the max values for scaling
    const maxX = 512 //Math.max(...coordinates.map(point => point[0]));
    const maxY = 512 //Math.max(...coordinates.map(point => point[1]), 512);

    // Draw grid and axis values
    const gridSpacingX = graphWidth / 10;
    const gridSpacingY = graphHeight / 10;

    // ctx.strokeStyle = '#d55';
    ctx.strokeStyle = cssColor(ctx, "graticule");
    ctx.lineWidth = 1;
    ctx.font = '12px Arial';
    ctx.fillStyle = cssColor(ctx, "legend");

    for (let i = 0; i <= 10; i++) {
        // Vertical grid lines and x-axis values
        const x = wmargin + i * gridSpacingX;
        ctx.beginPath();
        ctx.moveTo(x, hmargin);
        ctx.lineTo(x, hmargin + graphHeight);
        ctx.stroke();
        // const xValue = Math.round((i / 10) * maxX) / hMult;
        const xValue = Math.round(i * hMult);
        ctx.fillText(xValue, x - 10, hmargin + graphHeight + 15);

        // Horizontal grid lines and y-axis values
        const y = hmargin + i * gridSpacingY;
        ctx.beginPath();
        ctx.moveTo(wmargin, y);
        ctx.lineTo(wmargin + graphWidth, y);
        ctx.stroke();
        // const yValue = Math.round(maxY - (i / 10) * maxY) / vMult;
        const yValue = Math.round((5 - i) * vMult);
        ctx.fillText(yValue, wmargin - 40, y + 3); // Adjusted to avoid collision with vertical legend
    }

    // Draw x and y legends outside the graph
    ctx.font = '16px Arial';
    ctx.fillStyle = cssColor(ctx, "legend");

    // X-axis legend
    const xLegend = `${hMult}${hLegend} / div`
    // TODO calculate actual legend width
    ctx.fillText(xLegend, (canvas.width / 2) - 30, canvas.height - 10);

    // Y-axis legend (rotated 90 degrees)
    const yLegend = `${vMult}${vLegend} / div`
    ctx.save();
    ctx.translate(20, (canvas.height / 2) + 25); // Move to the left of the canvas
    ctx.rotate(-Math.PI / 2); // Rotate 90 degrees counterclockwise
    ctx.fillText(yLegend, 0, 0);
    ctx.restore();

    // Draw the line chart
    ctx.beginPath();
    ctx.moveTo(
        wmargin + (coordinates[0][0] / maxX) * graphWidth,
        hmargin + graphHeight - (coordinates[0][1] / maxY) * graphHeight
    );

    for (let i = 1; i < coordinates.length; i++) {
        const x = wmargin + (coordinates[i][0] / maxX) * graphWidth;
        const y = hmargin + graphHeight - (coordinates[i][1] / maxY) * graphHeight;
        ctx.lineTo(x, y);
    }

    ctx.lineCap = 'round';
    ctx.lineJoin = 'round';
    ctx.strokeStyle = cssColor(ctx, "traceedge0");
    ctx.lineWidth = cssWidth(ctx, "traceedge0");
    ctx.stroke();
    ctx.strokeStyle = cssColor(ctx, "traceedge1");
    ctx.lineWidth = cssWidth(ctx, "traceedge1");
    ctx.stroke();

    ctx.strokeStyle = cssColor(ctx, "trace");
    ctx.lineWidth = cssWidth(ctx, "trace");
    ctx.stroke();
}

// Function to periodically check for updates
async function updateChart() {
    const data = await fetchData();
    if (JSON.stringify(data) !== JSON.stringify(lastData)) {
        lastData = data;
        drawChart(data);
    }
}

// Event listener for the refresh button
refreshButton.addEventListener('click', async () => {
    const data = await fetchData();
    if (data) {
        lastData = data;
        drawChart(data);
    }
});


updateChart();

// Start periodic updates
// setInterval(updateChart, 5000); // Check every 5 seconds
// TODO add slider for refresh interval


// Update CSS variables dynamically based on slider values
const graticuleSlider = document.getElementById('graticule-slider');
const traceSlider = document.getElementById('trace-slider');

graticuleSlider.addEventListener('input', () => {
    // TODO shadow when too bright?
    const alpha = graticuleSlider.value;
    document.documentElement.style.setProperty('--graticule-color', `rgba(255, 123, 0, ${alpha})`);
    drawChart(lastData);
});

traceSlider.addEventListener('input', () => {
    const val = traceSlider.value;
    // widths
    // can't load from css as that will increase
    //    var w = cssWidth(ctx, "trace"), w1 = cssWidth(ctx, "traceedge1"), w0 = cssWidth(ctx, "traceedge0");
    var w = 1, w1 = 3, w0 = 5;
    // colors to override for primary trace (others do not change)
    var red = 17, blue = 0;

    var alpha = val;
    if (val > 1) {
        alpha = 1;
        red = Math.min(255, 17 + (val - 1) * 255);
        blue = Math.min(255, (val - 1) * 255);
        // w0 += 1;
        w *= val; w1 *= val; w0 *= val;
        // w += 1; w1 += 1; w0 += 1;
        // if (val > 1.25) { w += 1; w1 += 1; w0 += 1; }
        // if (val > 1.5) { w += 1; w1 += 1; w0 += 1; }
        // if (val > 1.75) { w += 1; w1 += 1; w0 += 1; }
        if (val > 2) {
            w *= val - 1; w1 *= val - 1; w0 *= val - 1;
        }
    }
    document.documentElement.style.setProperty('--trace-width', w);
    document.documentElement.style.setProperty('--traceedge1-width', w1);
    document.documentElement.style.setProperty('--traceedge0-width', w0);
    const color = `rgba(${red}, 255, ${blue}, ${alpha})`;
    //document.documentElement.style.setProperty('--trace-color', `rgba(17, 255, 0, ${alpha})`);
    document.documentElement.style.setProperty('--trace-color', color);
    // document.documentElement.style.setProperty('--traceedge1-color', `rgba(17, 255, 0, ${val * 0.5})`);
    // document.documentElement.style.setProperty('--traceedge0-color', `rgba(17, 255, 0, ${val * 0.25})`);
    document.documentElement.style.setProperty('--traceedge1-color', `rgba(17, 255, 0, ${val * 0.5})`);
    document.documentElement.style.setProperty('--traceedge0-color', `rgba(17, 255, 0, ${val * 0.25})`);
    drawChart(lastData);
});

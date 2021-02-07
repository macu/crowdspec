const VueLoaderPlugin = require('vue-loader/lib/plugin');
const MiniCssExtractPlugin = require("mini-css-extract-plugin");

module.exports = {
	mode: 'production',
	stats: 'errors-warnings',
	entry: [
		'./src/app.js',
	],
	output: {
		filename: 'compiled.js',
		path: __dirname + '/js',
	},
	optimization: {
		minimize: true,
	},
	performance: {
		hints: 'warning',
		maxEntrypointSize: 250000, // JS output 250 kB
		maxAssetSize: 250000, // CSS output 250 kB
	},
	externals: {
		'jquery': 'jQuery',
		'vue': 'Vue',
		'vuex': 'Vuex',
		'vue-router': 'VueRouter',
		'dragula': 'dragula',
		'moment': 'moment',
	},
	module: {
		rules: [
			{
				test: /\.vue$/,
				loader: 'vue-loader',
			},
			{
				test: /\.js$/,
				exclude: /(node_modules|bower_components)/,
				use: {
					loader: 'babel-loader',
					options: {
						presets: [
							['@babel/preset-env', {targets: '>1%'}],
						],
					},
				},
			},
			{
				test: /\.s?css$/,
				use: [
					MiniCssExtractPlugin.loader, // add support for `import 'file.scss';` in JS
					{
						loader: 'css-loader',
						options: {
							url: false,
						},
					},
					{
						loader: 'sass-loader',
						options: {
							sassOptions: {
								includePaths: [
									//__dirname + '/bower_components/bootstrap-sass/assets/stylesheets',
								],
							},
						},
					},
				],
			},
		],
	},
	plugins: [
		new VueLoaderPlugin(),
		new MiniCssExtractPlugin({
			// Output destination for compiled CSS
			filename: '../css/compiled.css',
		}),
	],
};

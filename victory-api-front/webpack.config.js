// webpack.config.js
var HtmlWebpackPlugin = require('html-webpack-plugin');
var path = require("path");
var npmRoot = path.resolve(__dirname + "/node_modules");
var nodeModulesDir = path.resolve(__dirname, './node_modules');
var webpack = require("webpack");
console.log("npmRoot", npmRoot);

module.exports = {
    // multiple entry points building different apps, admin and js front-end
    devtool: 'inline-source-map',
    entry: {
        admin: './templates/js/admin/admin.js'
    },
    output: {
        // path: __dirname + "/docs",
        filename: '[name].js'
    },
    module: {
        loaders: [
            {test: /\.html$/, loader: 'html-loader'},
            {test: /\.js$/, use: [
                    {loader: 'ng-annotate-loader'},
                    {loader:'babel-loader', query: {presets: ['env']}}
                ],
                exclude: [nodeModulesDir]
            },
        ]
    },
    resolve: {
        // you can now require('file') instead of require('file.coffee')
        extensions: ['.js', '.json', '.coffee', '.html']
    },
    plugins: [
        new webpack.IgnorePlugin(/angular/)
        // new HtmlWebpackPlugin({
        //     template: path.join(__dirname, 'index.html'),
        //     inject: true,
        //     hash: true,
        //     filename: 'index.html',
        // }),
    ],
};

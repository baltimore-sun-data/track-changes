const CleanWebpackPlugin = require("clean-webpack-plugin");
const ExtractTextPlugin = require("extract-text-webpack-plugin");
const ManifestPlugin = require("webpack-manifest-plugin");
const MinifyPlugin = require("babel-minify-webpack-plugin");
const path = require("path");
const webpack = require("webpack");

module.exports = (env = {}) => {
  let namePattern = env.production ? "[name].[chunkhash]" : "[name]";

  const extractSass = new ExtractTextPlugin({
    filename: `css/${namePattern}.css`
  });

  return {
    context: path.resolve(__dirname, "src"),
    entry: {
      main: "./scss/app.scss",
      index: "./js/index.js",
      listing: "./js/listing.js",
      vendor: [
        "es6-string-html-template",
        "moment",
        "tinysort",
        "./js/storage.js"
      ]
    },
    output: {
      filename: `js/${namePattern}.js`,
      path: path.resolve(__dirname, "assets/static")
    },
    module: {
      rules: [
        {
          test: /\.js$/,
          exclude: /node_modules/,
          use: {
            loader: "babel-loader",
            options: {
              presets: ["es2016"]
            }
          }
        },
        {
          test: /\.scss$/,
          use: extractSass.extract({
            use: [
              {
                loader: "css-loader"
              },
              {
                loader: "postcss-loader",
                options: {
                  ident: "postcss",
                  plugins: loader => [
                    require("postcss-import")({ root: loader.resourcePath }),
                    require("autoprefixer")(),
                    require("cssnano")()
                  ]
                }
              },
              {
                loader: "sass-loader"
              }
            ],
            // use style-loader in development
            fallback: "style-loader"
          })
        }
      ]
    },
    plugins: [
      new CleanWebpackPlugin(["assets"]),
      new webpack.HashedModuleIdsPlugin(),
      new webpack.optimize.CommonsChunkPlugin({
        name: "vendor"
      }),
      new webpack.optimize.CommonsChunkPlugin({
        name: "runtime"
      }),
      new ManifestPlugin({
        fileName: "../manifest.json",
        publicPath: "/static/"
      }),
      new MinifyPlugin(),
      new webpack.IgnorePlugin(/^\.\/locale$/, /moment$/),
      extractSass
    ]
  };
};

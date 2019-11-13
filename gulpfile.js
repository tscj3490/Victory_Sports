/* Copyright (C) NEEDA FZ LLC - All Rights Reserved
 * Unauthorized copying of this file, via any medium is strictly prohibited.
 * Proprietary and confidential.
 */
var gulp				=	require('gulp'),
	less					=	require('gulp-less'),
	browserSync		=	require('browser-sync'),
	concat				=	require('gulp-concat'),
	uglify				=	require('gulp-uglifyjs'),
	cssnano				=	require('gulp-cssnano'),
	rename				=	require('gulp-rename'),
	del						=	require('del'),
	imagemin			=	require('gulp-imagemin'),
	pngquant			=	require('imagemin-pngquant'),
	cache					=	require('gulp-cache'),
	autoprefixer	=	require('gulp-autoprefixer'),
	uncss					=	require('gulp-uncss'),
	sourcemaps	=	require('gulp-sourcemaps'),
    concatCss			=	require('gulp-concat-css'),
	gulpIncludeTemplate		=	require("gulp-include-template"),
    babel = require("gulp-babel");
var webpack = require('webpack-stream');
var child = require('child_process');
var exec = child.exec;
var util = require('gulp-util');
var notifier = require('node-notifier');
var plumber = require('gulp-plumber');
var sync = require('gulp-sync')(gulp).sync;
// var reload = require('gulp-livereload');

// adding typescript support
var ts = require('gulp-typescript');

/* ----------------------------------------------------------------------------
 * Locals
 * ------------------------------------------------------------------------- */

/* Application server */
var server = null;

var globalErr = undefined;

var runCommandOnChange = process.argv.slice(3);
if (runCommandOnChange.length >= 1) {
   runCommandOnChange = runCommandOnChange[0].replace("--cmd=","");
}

/* ----------------------------------------------------------------------------
 * Overrides
 * ------------------------------------------------------------------------- */

/*
 * Override gulp.src() for nicer error handling.
 */
var src = gulp.src;
gulp.src = function() {
    return src.apply(gulp, arguments)
        .pipe(plumber(function(error) {
                globalErr = error;
                util.log(util.colors.red(
                    'Error (' + error.plugin + '): ' + error.message
                ));
                notifier.notify({
                    title: 'Error (' + error.plugin + ')',
                    message: error.message.split('\n')[0]
                });
                this.emit('end');
            })
        );
};

/* ----------------------------------------------------------------------------
 * Assets pipeline
 * ------------------------------------------------------------------------- */

/*
 * Build assets.
 */
gulp.task('assets:build', [
    'assets:build:stylesheets',
    'assets:build:javascript'
    // 'assets:modernizr',
    // 'assets:views'
]);

gulp.task('assets:build:stylesheets', function(){
	return gulp.src('templates/less/*.less')
		.pipe( less() )
		/*.pipe( concatCss('css/main.css') )
		.pipe( uncss({
			html: ['app/index.html', 'resturants.html']
		}) )*/
		.pipe( autoprefixer(['last 2 versions'], { cascade: true }) )
		// .pipe( cssnano() )
		// .pipe( rename({suffix: '.min'}) )
		.pipe( gulp.dest('public/static/css') );
		// .pipe( browserSync.reload({ stream: true }) );
});

/*
 * Build javascripts from Bower components and source.
 */
gulp.task('assets:build:javascript', function() {
    return gulp.src("templates/js/admin/admin.js")
        .pipe(webpack({
            config: require('./webpack.config.js')
        }))
        .pipe(gulp.dest("public/static/js/"));
        // .pipe(sourcemaps.init())
        // .pipe(babel())
        // .pipe(concat("admin.js"))
        // .pipe(sourcemaps.write("."))
        // .pipe(gulp.dest("public/static/js/"));
    //
    // nothing to build yet
    //
    // return gulp.src([
    //     /* Your JS dependencies via bower_components */
    //     /* Your JS libraries */
    // ]).pipe(gulpif(args.sourcemaps, sourcemaps.init()))
    //     .pipe(concat('application.js'))
    //     .pipe(gulpif(args.sourcemaps, sourcemaps.write()))
    //     .pipe(gulpif(args.production, uglify()))
    //     .pipe(gulp.dest('public/javascripts/'))
    //     .pipe(reload());

    // return gulp.src('templates/ts/*.ts')
    //     .pipe(sourcemaps.init()) // This means sourcemaps will be generated
    //     .pipe(ts({
    //         noImplicitAny: true
    //         // outFile: 'admin.js'
    //     }))
    //     .pipe(sourcemaps.write('.', { includeContent: false, sourceRoot: '../templates/ts' }))
    //     .pipe(gulp.dest('public/static/js'));
});

/*
 * Watch assets for changes and rebuild on the fly.
 */
gulp.task('assets:watch', function() {

    /* Rebuild stylesheets on-the-fly */

    gulp.watch([
        'templates/less/*.less',
    ], ['assets:build:stylesheets', (function(){setTimeout(browserSync.reload,500);}) ]);

    // /* Rebuild javascripts on-the-fly */
    gulp.watch([
        'templates/js/**/*.js',
        'templates/emails/*.html',
        'templates/html/*.html'
    ], ['assets:build:javascript', (function(){setTimeout(browserSync.reload,500);}) ]);

    // /* Minify views on-the-fly */
    // gulp.watch([
    //     'views/**/*.tmpl'
    // ], ['assets:views']);

    // gulp.watch([
    //     'public/css/**/*.css',
    //     'templates/html/*.html'
    // ], browserSync.reload({ stream: false }));

});

/* ----------------------------------------------------------------------------
 * Application server
 * ------------------------------------------------------------------------- */

/*
 * Build and run application server.
 */
var executableName = 'victory-frontend';
var mainGoFile = 'main.go';

gulp.task('server:build', function() {
    // var build = child.spawnSync('go', ['install']);
    var build = child.spawnSync('go', ['build', '-o', executableName, mainGoFile]);
    if (build != undefined && build.stderr != undefined && build.stderr.length) {
        var lines = build.stderr.toString()
            .split('\n').filter(function(line) {
                return line.length
            });
        for (var l in lines)
            util.log(util.colors.red(
                'Error (go install): ' + lines[l]
            ));
        var err = {
            title: 'Error (go install)',
            message: lines
        };
        globalErr = err;
        notifier.notify(err);
    } else {
        console.log("server:build finished");
        globalErr = undefined;
    }
    return build;
});
/*
 * Restart application server.
 */
gulp.task('server:spawn', function() {
    if (server)
        server.kill();

    if (globalErr !== undefined) {
        console.log("server:spawn - global error set");
        return
    }
    /* Spawn application server */
    server = child.spawn('./'+executableName);

    /* Trigger reload upon server start */
    server.stdout.once('data', function() {
        // reload.reload('/');
        setTimeout(browserSync.reload, 500);
    });

    /* Pretty print server log output */
    server.stdout.on('data', function(data) {
        var lines = data.toString().split('\n')
        for (var l in lines)
            if (lines[l].length)
                util.log(lines[l]);
    });

    /* Print errors to stdout */
    server.stderr.on('data', function(data) {
        process.stdout.write(data.toString());
    });
});
/*
 * Watch source for changes and test the application server.
 */

gulp.task('server:test', function() {

    console.log("server:watch - setting up test watcher - ", runCommandOnChange);
    gulp.watch([
        // '*/**/*.go',
        // '*/**/**/*.go',
        // '*/**/**/**/*.go'
        '*/*/*.go',
        // 'libs/gosportmonks/*.go'
    ], {
        delay: 1000
    }, function() {


        console.log("running test command");
        exec(runCommandOnChange, function (err, stdout, stderr) {
            var lines = stdout.toString().split('\n');
            for (var l in lines)
                if (lines[l].length)
                    util.log(lines[l]);
            console.log("stdout:",stdout);
            console.log('stderr',stderr);
        });

        // server = child.spawn('./'+ runCommandOnChange);
        //
        // /* Pretty print server log output */
        // server.stdout.on('data', function(data) {
        //     var lines = data.toString().split('\n')
        //     for (var l in lines)
        //         if (lines[l].length)
        //             util.log(lines[l]);
        // });

        // done();
    });
});


gulp.task('server:watch', function() {
    /* Rebuild and restart application server */
    console.log("server:watch - setting up build watcher");
    gulp.watch([
        '*/**/*.go',
    ], sync([
        'server:build',
        'server:spawn'
    ], 'server'));
});



/*---------- HTML ----------*/
// gulpIncludeTemplate.config('base', 'app/html/');
// gulp.task('html', function(){
// 	return gulp.src(['app/html/*.html', '!app/html/__*.html'])
// 		.pipe(gulpIncludeTemplate())
// 		.pipe(gulp.dest("app/"));
// });

/*---------- Watch ----------*/
gulp.task('watch', function(){
	// gulp.watch('templates/less/*.less', ['less']);
	// gulp.watch('public/css/**/*.css', browserSync.reload({ stream: true }));
	// gulp.watch('app/html/*.html', ['html'])
	// gulp.watch('templates/*.html', browserSync.reload);
    gulp.start([
        'build',
        'server:spawn',
        'server:watch',
        'assets:watch',
        'browser-sync'
    ]);
});

/*---------- BrowserSync ----------*/
gulp.task('browser-sync', function(){
    browserSync.init({
        localOnly: true,
        //host: "127.0.0.1",
        host: "192.168.1.144",
        proxy: 'localhost:8080',
        serveStatic: [{
            route: '/public',
            dir: 'public'
        }],
        notify: false
    })
});

/*===============================
------------- BUILD -------------
===============================*/

gulp.task('clean', function() {
	return del.sync('dist');
});

/*
 * Build assets and application server.
 */
gulp.task('build', [
    'assets:build',
    //  UNCOMMENT WHEN SERVER IS BEING WORKED ON AS WELL
     'server:build'
], function(done){
    console.log("building finished - reloading browser");
    done();
});

/*
 * Build assets by default.
 */
gulp.task('default', ['build']);


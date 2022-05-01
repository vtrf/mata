use clap::{arg, Arg, Command};
use gray_matter::{engine::YAML, Matter};
use platform_dirs::AppDirs;
use reqwest::{header, StatusCode};
use serde::{Deserialize, Serialize};
use std::{
    fs::{self, File},
    process,
};

const NAME: &str = env!("CARGO_PKG_NAME");
const VERSION: &str = env!("CARGO_PKG_VERSION");

#[derive(Serialize, Deserialize)]
struct Config {
    endpoint: String,
    key: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct PostsBaseResponse {
    ok: bool,
    error: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
struct PostsPatchRequest {
    title: Option<String>,
    slug: Option<String>,
    body: Option<String>,
    published_at: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]
struct PostsPatchResponse {
    ok: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
    slug: String,
    url: String,
}

#[derive(Deserialize, Debug)]
struct FrontMatter {
    title: Option<String>,
    published_at: Option<String>,
    slug: Option<String>,
}

#[derive(Serialize, Deserialize, Debug)]

struct Post {
    title: String,
    slug: String,
    body: String,
    published_at: String,
    url: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct PostsGetResponse {
    ok: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
    slug: String,
    title: String,
    body: String,
    published_at: String,
    url: String,
}

#[derive(Serialize, Deserialize, Debug)]
struct PostsListResponse {
    ok: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
    post_list: Vec<Post>,
}

fn cli() -> Command<'static> {
    Command::new(NAME)
        .about("CLI tool for mataroa.blog")
        .version(VERSION)
        .subcommand_required(true)
        .subcommand(Command::new("version").about("Show version"))
        .subcommand(Command::new("init").about("Initialize mata"))
        .subcommand(
            Command::new("posts")
                .about("Manage posts")
                .subcommand(
                    Command::new("delete")
                        .about("Delete a post")
                        .arg(arg!(<SLUG> "Post slug"))
                        .arg_required_else_help(true),
                )
                .subcommand(
                    Command::new("edit")
                        .about("Delete a post")
                        .arg(Arg::new("slug").help("Post slug"))
                        .arg_required_else_help(true),
                )
                .subcommand(
                    Command::new("get")
                        .about("Delete a post")
                        .arg(Arg::new("slug").help("Post slug"))
                        .arg_required_else_help(true),
                )
                .subcommand(Command::new("list").about("Delete a post"))
                .subcommand(
                    Command::new("update")
                        .about("Delete a post")
                        .args(&[
                            Arg::new("slug").help("Post slug"),
                            Arg::new("file").help("Post markdown file"),
                        ])
                        .arg_required_else_help(true),
                ),
        )
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let matches = cli().get_matches();

    let app_dirs = AppDirs::new(Some(NAME), true).unwrap();
    let config_file_path = app_dirs.config_dir.join("config.json");

    match matches.subcommand() {
        Some(("version", _)) => {
            println!("{} {}", NAME, VERSION);
        }
        Some(("init", _)) => {
            if config_file_path.exists() {
                println!("{} already exists", config_file_path.as_path().display());
                process::exit(1);
            } else {
                let default_cfg = Config {
                    endpoint: "https://mataroa.blog/api".into(),
                    key: "your-api-key".into(),
                };

                serde_json::to_writer_pretty(
                    &File::create(config_file_path).unwrap(),
                    &default_cfg,
                )
                .unwrap();

                // TODO: print config_file_path too
                println!("mata initilized")
            };
        }
        Some(("posts", sub_matches)) => {
            let stash_command = sub_matches.subcommand().unwrap();

            let config: Config = serde_json::from_reader(File::open(config_file_path)?)
                .expect("couldn't open config file");

            let mut headers = header::HeaderMap::new();
            headers.insert(
                header::AUTHORIZATION,
                header::HeaderValue::from_str(format!("Bearer {}", config.key).as_str())?,
            );

            let client = reqwest::Client::builder()
                .default_headers(headers)
                .build()?;

            match stash_command {
                ("delete", args) => {
                    let slug = args.value_of("SLUG").unwrap();

                    let response = client
                        .delete(format!("{}/posts/{}", config.endpoint, slug))
                        .send()
                        .await?;

                    match response.status() {
                        StatusCode::NOT_FOUND => {
                            println!("post '{}' not found", slug);
                            process::exit(1);
                        }
                        _ => {
                            let resp = response.json::<PostsBaseResponse>().await?;
                            if !resp.ok {
                                println!("{}", resp.error.unwrap());
                                process::exit(1);
                            }

                            println!("post '{}' deleted successfully", slug);
                        }
                    }
                }
                ("get", args) => {
                    let slug = args.value_of("SLUG").unwrap();

                    let response = client
                        .get(format!("{}/posts/{}", config.endpoint, slug))
                        .send()
                        .await?;

                    match response.status() {
                        StatusCode::NOT_FOUND => {
                            println!("post '{}' not found", slug);
                            process::exit(1);
                        }
                        _ => {
                            let resp = response.json::<PostsGetResponse>().await?;
                            if !resp.ok {
                                println!("{}", resp.error.unwrap());
                                process::exit(1);
                            }

                            let pretty = serde_json::to_string_pretty(&resp).unwrap();
                            println!("{}", pretty)
                        }
                    }
                }
                ("list", _) => {
                    let response = client
                        .get(format!("{}/posts", config.endpoint))
                        .send()
                        .await?;

                    let resp = response.json::<PostsListResponse>().await?;
                    if !resp.ok {
                        println!("{}", resp.error.unwrap());
                        process::exit(1);
                    }

                    let pretty = serde_json::to_string_pretty(&resp).unwrap();
                    println!("{}", pretty)
                }
                ("update", args) => {
                    let slug = args.value_of("slug").unwrap();
                    let file = args.value_of("file").unwrap();

                    let content = fs::read_to_string(file).expect("couldn't read file");

                    let matter = Matter::<YAML>::new();
                    let result = matter
                        .parse_with_struct::<FrontMatter>(content.as_str())
                        .unwrap();

                    let post = PostsPatchRequest {
                        published_at: result.data.published_at,
                        title: result.data.title,
                        slug: result.data.slug,
                        body: Some(result.content),
                    };

                    let response = client
                        .patch(format!("{}/posts/{}", config.endpoint, slug))
                        .json(&post)
                        .send()
                        .await?;

                    match response.status() {
                        StatusCode::NOT_FOUND => {
                            println!("post '{}' not found", slug);
                            process::exit(1);
                        }
                        _ => {
                            let resp = response.json::<PostsPatchResponse>().await?;
                            if !resp.ok {
                                println!("{}", resp.error.unwrap());
                                process::exit(1);
                            }

                            let pretty = serde_json::to_string_pretty(&resp).unwrap();
                            println!("{}", pretty)
                        }
                    }
                }
                _ => unreachable!(),
            }
        }
        _ => unreachable!(),
    }

    Ok(())
}

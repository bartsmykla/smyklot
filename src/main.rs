#![feature(allocator_api)]

use serenity::async_trait;
use serenity::client::{Client, Context, EventHandler};
use serenity::model::channel::Message;
use serenity::framework::standard::{
    StandardFramework,
    CommandResult,
    macros::{
        command,
        group,
    },
};

use std::env;
use serenity::http::Http;
use std::collections::HashSet;
use serenity::model::id::GuildId;
use std::alloc::Global;

#[group]
#[commands(ping, pong, pif, paf, do_you_know, version)]
struct General;

struct Handler;

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, ctx: Context, msg: Message) {
        let bot_user_ud = ctx.cache.current_user_id().await;
        if msg.content == format!("<@!{}> po ile schab?", bot_user_ud.to_string()) {
            if msg.author.name == "bartsmykla" {
                msg.reply(ctx, "dla Ciebie dycha").await;
            } else {
                msg.reply(ctx, "nie stać cię").await;
            }
        }
    }
}

#[tokio::main]
async fn main() {    // Login with a bot token from the environment
    let token = env::var("DISCORD_TOKEN").expect("token");

    let http = Http::new_with_token(&token);

    // We will fetch your bot owners and id
    let (_owners, _bot_id) = match http.get_current_application_info().await {
        Ok(info) => {
            println!("{:?}", info);
            let mut owners = HashSet::new();
            owners.insert(info.owner.id);

            (owners, info.id)
        },
        Err(why) => panic!("Could not access application info: {:?}", why),
    };

    println!("{}", _bot_id);

    let framework = StandardFramework::new()
        .configure(|c| c.prefix(format!("<@!{}>", _bot_id).as_str()).with_whitespace(true)) // set the bot prefix to "!"
        .group(&GENERAL_GROUP);

    let mut client = Client::builder(token)
        .event_handler(Handler)
        .framework(framework)
        .await
        .expect("Error creating client");

    // start listening for events by starting a single shard
    if let Err(why) = client.start().await {
        println!("An error occurred while running the client: {:?}", why);
    }
}

enum PingType {
    Ping,
    Pong,
    Pif,
    Paf,
}

impl PingType {
    fn reply(self) -> &'static str {
        match self {
            PingType::Ping => "Pong",
            PingType::Pong => "Ping",
            PingType::Pif => "Paf",
            PingType::Paf => "Pif",
        }
    }
}

#[command]
async fn ping(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, PingType::Ping.reply()).await?;

    Ok(())
}

#[command]
async fn pong(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, PingType::Pong.reply()).await?;

    Ok(())
}

#[command]
async fn pif(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, PingType::Pif.reply()).await?;

    Ok(())
}

#[command]
async fn version(ctx: &Context, msg: &Message) -> CommandResult {
    let version = env::var("SMYKLOT_VERSION");
    
    let message = match version {
        Ok(v) if v != "{{version}}" => v,
        _ => String::from("¯\\_(ツ)_/¯")
    };

    msg.reply(ctx, message).await?;

    Ok(())
}

#[command]
async fn paf(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, PingType::Paf.reply()).await?;

    Ok(())
}

#[command("znasz")]
async fn do_you_know(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "pierwsze słyszę").await?;

    Ok(())
}

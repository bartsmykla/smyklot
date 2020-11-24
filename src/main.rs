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

#[group]
#[commands(ping, pong, pif, paf)]
struct General;

struct Handler;

#[async_trait]
impl EventHandler for Handler {}

#[tokio::main]
async fn main() {
    let framework = StandardFramework::new()
        .configure(|c| c.prefix("!")) // set the bot prefix to "!"
        .group(&GENERAL_GROUP);

    // Login with a bot token from the environment
    let token = env::var("DISCORD_TOKEN").expect("token");
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
async fn paf(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, PingType::Paf.reply()).await?;

    Ok(())
}


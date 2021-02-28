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
use serenity::http::Http;
use log::{info, error};

use std::env;
use std::collections::HashSet;

#[group]
#[commands(ping, pong, pif, paf, do_you_know, version)]
struct General;

struct Handler;

#[async_trait]
impl EventHandler for Handler {
    async fn message(&self, ctx: Context, msg: Message) {
        let bot_user_ud = ctx.cache.current_user_id().await;
        
        if msg.content == format!("<@!{}> po ile schab?", bot_user_ud.to_string()) {
            let message = if msg.author.name == "bartsmykla" {
                "dla Ciebie dycha"
            } else {
                "nie stać cię"
            };
            
            if let Err(e) = msg.reply(ctx, message).await {
                error!("Error when tried to send a message: {}", e)
            }
        }
    }
}

#[tokio::main]
async fn main() {
    env_logger::init();
    
    let token = env::var("DISCORD_TOKEN").expect("token");

    let http = Http::new_with_token(&token);

    // We will fetch your bot owners and id
    let (_owners, bot_id) = match http.get_current_application_info().await {
        Ok(info) => {
            info!("{:?}", info);
            let mut owners = HashSet::new();
            owners.insert(info.owner.id);

            (owners, info.id)
        },
        Err(why) => panic!("Could not access application info: {:?}", why),
    };

    info!("bot id: {}", bot_id);
    
    let bot_prefix = format!("<@!{}>", bot_id);
    let prefixes = vec![bot_prefix.as_str(), "!"];

    let framework = StandardFramework::new()
        .configure(|c| c.prefixes(prefixes).with_whitespace(true))
        .group(&GENERAL_GROUP);

    let mut client = Client::builder(token)
        .event_handler(Handler)
        .framework(framework)
        .await
        .expect("Error creating client");

    // start listening for events by starting a single shard
    if let Err(why) = client.start().await {
        error!("An error occurred while running the client: {:?}", why);
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

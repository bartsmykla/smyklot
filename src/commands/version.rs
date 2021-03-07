use serenity::{
    prelude::*,
    model::{
        channel::{Message},
    },
    framework::standard::{
        CommandResult,
        macros::*,
    },
};

use crate::Config;

#[command]
async fn version(ctx: &Context, msg: &Message) -> CommandResult {
    let config_lock = ctx.data.read().await
        .get::<Config>()
        .expect("Missing Config in Context")
        .clone();

    let config = config_lock.read().await;
    let version = config.version.as_str();

    let message = match version {
        "{{version}}" | "" => "¯\\_(ツ)_/¯",
        _ => version
    };

    msg.reply(ctx, message).await?;

    Ok(())
}

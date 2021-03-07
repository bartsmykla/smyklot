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

#[command]
#[aliases("znasz", "know")]
async fn do_you_know(ctx: &Context, msg: &Message) -> CommandResult {
    if msg.author.name == "zawiszaty" {
        msg.reply(ctx, "tobie nie powiem").await?;

        return Ok(());
    }

    msg.reply(ctx, "pierwsze słyszę").await?;

    Ok(())
}

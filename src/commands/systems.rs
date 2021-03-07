use serenity::{
    prelude::*,
    model::{
        channel::{Message},
    },
    framework::standard::{
        Args, CommandResult,
        macros::*,
    },
};

#[command]
#[description = "Sends opinion about macos."]
#[aliases("apple", "macos", "mac")]
#[bucket = "systems"]
async fn mac(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jak cię stać na ten szmelc").await?;

    Ok(())
}

#[command]
#[description = "Sends opinion about linux os."]
#[aliases("pingwinie", "ubuntu", "i3")]
#[bucket = "systems"]
async fn linux(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jedyne słuszne rozwiązanie! :sunglasses:").await?;

    Ok(())
}

#[command]
#[description = "Sends opinion about windows."]
#[aliases("winda", "windows 10", "windows vista", "windows xp")]
#[bucket = "systems"]
async fn windows(ctx: &Context, msg: &Message) -> CommandResult {
    msg.reply(ctx, "Jak zrestartujesz kompa to pogadamy").await?;

    Ok(())
}
